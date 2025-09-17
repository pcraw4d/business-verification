-- =============================================================================
-- FINANCIAL SERVICES KEYWORDS TEST SCRIPT
-- Task 3.2.6: Test financial services keywords for classification accuracy
-- =============================================================================
-- This script tests the financial services keywords to ensure they provide
-- >85% classification accuracy for financial businesses.
-- 
-- Test Categories:
-- 1. Banking Classification Tests
-- 2. Insurance Classification Tests  
-- 3. Investment Services Classification Tests
-- 4. Cross-Industry Differentiation Tests
-- 5. Edge Case Tests
-- =============================================================================

-- =============================================================================
-- TEST 1: BANKING CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 1: BANKING CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test banking business descriptions
WITH banking_tests AS (
    SELECT 
        'Chase Bank' as business_name,
        'Commercial banking services including deposits, loans, and credit services' as description,
        'Banking' as expected_industry,
        0.80 as min_confidence
    UNION ALL
    SELECT 
        'Wells Fargo',
        'Retail banking with checking accounts, savings accounts, and mortgage services',
        'Banking',
        0.80
    UNION ALL
    SELECT 
        'Bank of America',
        'Full-service financial institution offering commercial and retail banking',
        'Banking',
        0.80
    UNION ALL
    SELECT 
        'Credit Union of America',
        'Member-owned financial cooperative providing banking and lending services',
        'Banking',
        0.75
    UNION ALL
    SELECT 
        'Community Savings Bank',
        'Local community bank specializing in small business loans and personal banking',
        'Banking',
        0.75
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test banking classification accuracy' as test_purpose
FROM banking_tests;

-- =============================================================================
-- TEST 2: INSURANCE CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 2: INSURANCE CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test insurance business descriptions
WITH insurance_tests AS (
    SELECT 
        'State Farm Insurance' as business_name,
        'Property and casualty insurance including auto, home, and life insurance' as description,
        'Insurance' as expected_industry,
        0.75 as min_confidence
    UNION ALL
    SELECT 
        'Allstate Insurance',
        'Insurance company offering auto, home, and business insurance policies',
        'Insurance',
        0.75
    UNION ALL
    SELECT 
        'Progressive Insurance',
        'Auto insurance company with competitive rates and comprehensive coverage',
        'Insurance',
        0.75
    UNION ALL
    SELECT 
        'Blue Cross Blue Shield',
        'Health insurance provider offering medical, dental, and vision coverage',
        'Insurance',
        0.75
    UNION ALL
    SELECT 
        'Aetna Health Insurance',
        'Health insurance company providing medical coverage and wellness programs',
        'Insurance',
        0.75
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test insurance classification accuracy' as test_purpose
FROM insurance_tests;

-- =============================================================================
-- TEST 3: INVESTMENT SERVICES CLASSIFICATION TESTS
-- =============================================================================
SELECT 
    'TEST 3: INVESTMENT SERVICES CLASSIFICATION TESTS' as test_name,
    '' as spacer;

-- Test investment services business descriptions
WITH investment_tests AS (
    SELECT 
        'Fidelity Investments' as business_name,
        'Investment advisory services, wealth management, and retirement planning' as description,
        'Investment Services' as expected_industry,
        0.80 as min_confidence
    UNION ALL
    SELECT 
        'Charles Schwab',
        'Investment brokerage services, financial planning, and portfolio management',
        'Investment Services',
        0.80
    UNION ALL
    SELECT 
        'Vanguard Group',
        'Investment management company offering mutual funds and ETF products',
        'Investment Services',
        0.80
    UNION ALL
    SELECT 
        'Morgan Stanley',
        'Investment banking and wealth management services for high net worth clients',
        'Investment Services',
        0.80
    UNION ALL
    SELECT 
        'Edward Jones',
        'Financial advisor network providing investment advice and retirement planning',
        'Investment Services',
        0.75
)
SELECT 
    business_name,
    description,
    expected_industry,
    min_confidence,
    'Test investment services classification accuracy' as test_purpose
FROM investment_tests;

-- =============================================================================
-- TEST 4: CROSS-INDUSTRY DIFFERENTIATION TESTS
-- =============================================================================
SELECT 
    'TEST 4: CROSS-INDUSTRY DIFFERENTIATION TESTS' as test_name,
    '' as spacer;

-- Test that similar businesses are classified correctly
WITH differentiation_tests AS (
    SELECT 
        'Bank of New York Mellon' as business_name,
        'Investment banking and asset management services' as description,
        'Investment Services' as expected_industry,
        'Banking' as should_not_be,
        0.75 as min_confidence
    UNION ALL
    SELECT 
        'Goldman Sachs',
        'Investment banking, securities, and investment management',
        'Investment Services',
        'Banking',
        0.80
    UNION ALL
    SELECT 
        'JP Morgan Chase',
        'Investment banking and commercial banking services',
        'Banking',
        'Investment Services',
        0.75
    UNION ALL
    SELECT 
        'Prudential Financial',
        'Life insurance and investment management services',
        'Insurance',
        'Investment Services',
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
-- TEST 5: EDGE CASE TESTS
-- =============================================================================
SELECT 
    'TEST 5: EDGE CASE TESTS' as test_name,
    '' as spacer;

-- Test edge cases and ambiguous descriptions
WITH edge_case_tests AS (
    SELECT 
        'Financial Solutions Inc' as business_name,
        'Financial consulting and advisory services' as description,
        'Investment Services' as expected_industry,
        0.60 as min_confidence
    UNION ALL
    SELECT 
        'Money Management LLC',
        'Personal finance and money management services',
        'Investment Services',
        0.60
    UNION ALL
    SELECT 
        'Credit Solutions',
        'Credit repair and debt consolidation services',
        'Banking',
        0.60
    UNION ALL
    SELECT 
        'Risk Management Corp',
        'Risk assessment and insurance consulting services',
        'Insurance',
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

-- Analyze keyword coverage for each financial services industry
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
WHERE i.name IN ('Banking', 'Insurance', 'Investment Services') 
AND i.is_active = true
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.name;

-- =============================================================================
-- KEYWORD OVERLAP ANALYSIS
-- =============================================================================
SELECT 
    'KEYWORD OVERLAP ANALYSIS' as analysis_type,
    '' as spacer;

-- Check for keyword overlaps between financial services industries
WITH keyword_overlaps AS (
    SELECT 
        ik.keyword,
        COUNT(DISTINCT i.name) as industry_count,
        STRING_AGG(i.name, ', ' ORDER BY i.name) as industries
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name IN ('Banking', 'Insurance', 'Investment Services') 
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
WHERE i.name IN ('Banking', 'Insurance', 'Investment Services') 
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
WHERE i.name IN ('Banking', 'Insurance', 'Investment Services') 
AND i.is_active = true
GROUP BY i.id, i.name
ORDER BY i.name;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'FINANCIAL SERVICES KEYWORDS TEST SCRIPT COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Test Categories: 5 comprehensive test suites';
    RAISE NOTICE 'Test Cases: 20+ business classification scenarios';
    RAISE NOTICE 'Analysis: Keyword coverage, overlap, and performance validation';
    RAISE NOTICE 'Status: Ready for API testing and accuracy validation';
    RAISE NOTICE 'Next: Execute API tests to validate >85% classification accuracy';
    RAISE NOTICE '=============================================================================';
END $$;
