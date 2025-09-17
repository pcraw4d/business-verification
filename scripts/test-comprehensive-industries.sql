-- =============================================================================
-- COMPREHENSIVE INDUSTRY EXPANSION TESTING SCRIPT
-- =============================================================================
-- This script validates the comprehensive industry expansion and ensures
-- all 27 new industries were added correctly with proper configurations.
-- =============================================================================

-- =============================================================================
-- TEST 1: VERIFY ALL NEW INDUSTRIES EXIST
-- =============================================================================
DO $$
DECLARE
    industry_count INTEGER;
    expected_count INTEGER := 27;
    missing_industries TEXT[] := ARRAY[]::TEXT[];
    industry_name TEXT;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 1: VERIFYING ALL NEW INDUSTRIES EXIST';
    RAISE NOTICE '=============================================================================';
    
    -- Check each expected industry
    FOR industry_name IN 
        SELECT unnest(ARRAY[
            'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
            'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
            'Banking', 'Insurance', 'Investment Services', 'Fintech',
            'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
            'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
            'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
            'Software Development', 'Technology Services', 'Digital Services'
        ])
    LOOP
        IF NOT EXISTS (SELECT 1 FROM industries WHERE name = industry_name AND is_active = true) THEN
            missing_industries := array_append(missing_industries, industry_name);
        END IF;
    END LOOP;
    
    -- Count total new industries
    SELECT COUNT(*) INTO industry_count
    FROM industries
    WHERE is_active = true
    AND name IN (
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    -- Report results
    IF industry_count = expected_count AND array_length(missing_industries, 1) IS NULL THEN
        RAISE NOTICE '✅ SUCCESS: All % new industries found', industry_count;
    ELSE
        RAISE NOTICE '❌ FAILURE: Expected % industries, found %', expected_count, industry_count;
        IF array_length(missing_industries, 1) > 0 THEN
            RAISE NOTICE 'Missing industries: %', array_to_string(missing_industries, ', ');
        END IF;
    END IF;
END $$;

-- =============================================================================
-- TEST 2: VERIFY CATEGORY DISTRIBUTION
-- =============================================================================
DO $$
DECLARE
    traditional_count INTEGER;
    emerging_count INTEGER;
    hybrid_count INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 2: VERIFYING CATEGORY DISTRIBUTION';
    RAISE NOTICE '=============================================================================';
    
    SELECT 
        COUNT(CASE WHEN category = 'traditional' THEN 1 END),
        COUNT(CASE WHEN category = 'emerging' THEN 1 END),
        COUNT(CASE WHEN category = 'hybrid' THEN 1 END)
    INTO traditional_count, emerging_count, hybrid_count
    FROM industries
    WHERE is_active = true
    AND name IN (
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    RAISE NOTICE 'Traditional industries: %', traditional_count;
    RAISE NOTICE 'Emerging industries: %', emerging_count;
    RAISE NOTICE 'Hybrid industries: %', hybrid_count;
    
    IF traditional_count >= 20 AND emerging_count >= 5 THEN
        RAISE NOTICE '✅ SUCCESS: Good distribution of traditional and emerging industries';
    ELSE
        RAISE NOTICE '⚠️ WARNING: Unexpected category distribution';
    END IF;
END $$;

-- =============================================================================
-- TEST 3: VERIFY CONFIDENCE THRESHOLDS
-- =============================================================================
DO $$
DECLARE
    min_threshold DECIMAL;
    max_threshold DECIMAL;
    avg_threshold DECIMAL;
    invalid_count INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 3: VERIFYING CONFIDENCE THRESHOLDS';
    RAISE NOTICE '=============================================================================';
    
    SELECT 
        MIN(confidence_threshold),
        MAX(confidence_threshold),
        ROUND(AVG(confidence_threshold), 3),
        COUNT(CASE WHEN confidence_threshold < 0.70 OR confidence_threshold > 0.85 THEN 1 END)
    INTO min_threshold, max_threshold, avg_threshold, invalid_count
    FROM industries
    WHERE is_active = true
    AND name IN (
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    RAISE NOTICE 'Min confidence threshold: %', min_threshold;
    RAISE NOTICE 'Max confidence threshold: %', max_threshold;
    RAISE NOTICE 'Average confidence threshold: %', avg_threshold;
    RAISE NOTICE 'Invalid thresholds (outside 0.70-0.85): %', invalid_count;
    
    IF invalid_count = 0 AND min_threshold >= 0.70 AND max_threshold <= 0.85 THEN
        RAISE NOTICE '✅ SUCCESS: All confidence thresholds are within valid range';
    ELSE
        RAISE NOTICE '❌ FAILURE: Invalid confidence thresholds found';
    END IF;
END $$;

-- =============================================================================
-- TEST 4: VERIFY NO DUPLICATES
-- =============================================================================
DO $$
DECLARE
    duplicate_count INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 4: VERIFYING NO DUPLICATE INDUSTRIES';
    RAISE NOTICE '=============================================================================';
    
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT name, COUNT(*) as count
        FROM industries
        WHERE is_active = true
        GROUP BY name
        HAVING COUNT(*) > 1
    ) duplicates;
    
    IF duplicate_count = 0 THEN
        RAISE NOTICE '✅ SUCCESS: No duplicate industry names found';
    ELSE
        RAISE NOTICE '❌ FAILURE: % duplicate industry names found', duplicate_count;
    END IF;
END $$;

-- =============================================================================
-- TEST 5: VERIFY INDUSTRY DESCRIPTIONS
-- =============================================================================
DO $$
DECLARE
    empty_desc_count INTEGER;
    short_desc_count INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TEST 5: VERIFYING INDUSTRY DESCRIPTIONS';
    RAISE NOTICE '=============================================================================';
    
    SELECT 
        COUNT(CASE WHEN description IS NULL OR description = '' THEN 1 END),
        COUNT(CASE WHEN LENGTH(description) < 20 THEN 1 END)
    INTO empty_desc_count, short_desc_count
    FROM industries
    WHERE is_active = true
    AND name IN (
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    RAISE NOTICE 'Empty descriptions: %', empty_desc_count;
    RAISE NOTICE 'Short descriptions (<20 chars): %', short_desc_count;
    
    IF empty_desc_count = 0 AND short_desc_count = 0 THEN
        RAISE NOTICE '✅ SUCCESS: All industry descriptions are complete and descriptive';
    ELSE
        RAISE NOTICE '⚠️ WARNING: Some industry descriptions need improvement';
    END IF;
END $$;

-- =============================================================================
-- DETAILED INDUSTRY REPORT
-- =============================================================================
SELECT 
    'DETAILED INDUSTRY REPORT' as report_type,
    '' as spacer;

SELECT 
    category,
    name,
    confidence_threshold,
    CASE 
        WHEN confidence_threshold >= 0.80 THEN 'High'
        WHEN confidence_threshold >= 0.75 THEN 'Medium-High'
        WHEN confidence_threshold >= 0.70 THEN 'Medium'
        ELSE 'Low'
    END as confidence_level,
    LENGTH(description) as description_length
FROM industries
WHERE is_active = true
AND name IN (
    'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
    'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
    'Banking', 'Insurance', 'Investment Services', 'Fintech',
    'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
    'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
    'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
    'Software Development', 'Technology Services', 'Digital Services'
)
ORDER BY category, confidence_threshold DESC;

-- =============================================================================
-- SUMMARY STATISTICS
-- =============================================================================
SELECT 
    'SUMMARY STATISTICS' as summary_type,
    '' as spacer;

SELECT 
    COUNT(*) as total_new_industries,
    COUNT(CASE WHEN category = 'traditional' THEN 1 END) as traditional_industries,
    COUNT(CASE WHEN category = 'emerging' THEN 1 END) as emerging_industries,
    COUNT(CASE WHEN category = 'hybrid' THEN 1 END) as hybrid_industries,
    ROUND(MIN(confidence_threshold), 2) as min_confidence,
    ROUND(MAX(confidence_threshold), 2) as max_confidence,
    ROUND(AVG(confidence_threshold), 2) as avg_confidence
FROM industries
WHERE is_active = true
AND name IN (
    'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
    'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
    'Banking', 'Insurance', 'Investment Services', 'Fintech',
    'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
    'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
    'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
    'Software Development', 'Technology Services', 'Digital Services'
);

-- =============================================================================
-- FINAL VERIFICATION
-- =============================================================================
DO $$
DECLARE
    total_industries INTEGER;
    restaurant_industries INTEGER;
    new_industries INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'FINAL VERIFICATION';
    RAISE NOTICE '=============================================================================';
    
    SELECT COUNT(*) INTO total_industries FROM industries WHERE is_active = true;
    
    SELECT COUNT(*) INTO restaurant_industries
    FROM industries
    WHERE is_active = true
    AND name IN ('Restaurants', 'Fast Food', 'Food & Beverage', 'Fine Dining', 'Casual Dining', 'Quick Service', 'Catering', 'Food Trucks', 'Cafes & Coffee Shops', 'Bars & Pubs', 'Breweries', 'Wineries');
    
    SELECT COUNT(*) INTO new_industries
    FROM industries
    WHERE is_active = true
    AND name IN (
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    RAISE NOTICE 'Total active industries: %', total_industries;
    RAISE NOTICE 'Restaurant industries: %', restaurant_industries;
    RAISE NOTICE 'New industries added: %', new_industries;
    RAISE NOTICE 'Expected total: %', restaurant_industries + new_industries;
    
    IF total_industries >= 35 THEN
        RAISE NOTICE '✅ SUCCESS: Comprehensive industry expansion completed successfully';
        RAISE NOTICE 'Ready for keyword expansion (Task 3.2)';
    ELSE
        RAISE NOTICE '⚠️ WARNING: Industry count lower than expected';
    END IF;
END $$;
