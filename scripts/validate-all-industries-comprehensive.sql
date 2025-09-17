-- =============================================================================
-- COMPREHENSIVE INDUSTRY KEYWORD VALIDATION SCRIPT
-- Task 3.2: Validate all industries have adequate keyword coverage
-- =============================================================================
-- This script validates that all 39 industries in the database have adequate
-- keyword coverage for accurate classification (>85% accuracy target).
-- 
-- Validation Categories:
-- 1. Overall Industry Coverage Analysis
-- 2. Industry-Specific Keyword Coverage
-- 3. Keyword Weight Distribution Analysis
-- 4. Cross-Industry Keyword Overlap Analysis
-- 5. Performance and Quality Metrics
-- 6. Success Criteria Validation
-- =============================================================================

-- =============================================================================
-- VALIDATION 1: OVERALL INDUSTRY COVERAGE ANALYSIS
-- =============================================================================
SELECT 
    'VALIDATION 1: OVERALL INDUSTRY COVERAGE ANALYSIS' as validation_name,
    '' as spacer;

-- Check total industries and their keyword coverage
SELECT 
    COUNT(DISTINCT i.id) as total_industries,
    COUNT(DISTINCT CASE WHEN i.is_active = true THEN i.id END) as active_industries,
    COUNT(DISTINCT CASE WHEN ik.is_active = true THEN i.id END) as industries_with_keywords,
    COUNT(ik.id) as total_keywords,
    COUNT(CASE WHEN ik.is_active = true THEN 1 END) as active_keywords,
    ROUND(COUNT(ik.id)::decimal / COUNT(DISTINCT i.id), 1) as avg_keywords_per_industry
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.is_active = true;

-- =============================================================================
-- VALIDATION 2: INDUSTRY-SPECIFIC KEYWORD COVERAGE
-- =============================================================================
SELECT 
    'VALIDATION 2: INDUSTRY-SPECIFIC KEYWORD COVERAGE' as validation_name,
    '' as spacer;

-- Check keyword coverage for each industry
SELECT 
    i.name as industry_name,
    i.category,
    i.confidence_threshold,
    COUNT(ik.id) as keyword_count,
    ROUND(MIN(ik.weight), 3) as min_weight,
    ROUND(MAX(ik.weight), 3) as max_weight,
    ROUND(AVG(ik.weight), 3) as avg_weight,
    CASE 
        WHEN COUNT(ik.id) >= 50 THEN 'EXCELLENT'
        WHEN COUNT(ik.id) >= 30 THEN 'GOOD'
        WHEN COUNT(ik.id) >= 20 THEN 'ADEQUATE'
        WHEN COUNT(ik.id) >= 10 THEN 'POOR'
        ELSE 'CRITICAL'
    END as coverage_status
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.category, i.confidence_threshold
ORDER BY keyword_count DESC;

-- =============================================================================
-- VALIDATION 3: KEYWORD WEIGHT DISTRIBUTION ANALYSIS
-- =============================================================================
SELECT 
    'VALIDATION 3: KEYWORD WEIGHT DISTRIBUTION ANALYSIS' as validation_name,
    '' as spacer;

-- Analyze keyword weight distribution across all industries
SELECT 
    'Weight Distribution Summary' as analysis_type,
    COUNT(CASE WHEN ik.weight >= 0.90 THEN 1 END) as very_high_weight,
    COUNT(CASE WHEN ik.weight >= 0.80 AND ik.weight < 0.90 THEN 1 END) as high_weight,
    COUNT(CASE WHEN ik.weight >= 0.70 AND ik.weight < 0.80 THEN 1 END) as medium_high_weight,
    COUNT(CASE WHEN ik.weight >= 0.60 AND ik.weight < 0.70 THEN 1 END) as medium_weight,
    COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight < 0.60 THEN 1 END) as medium_low_weight,
    COUNT(CASE WHEN ik.weight < 0.50 THEN 1 END) as low_weight,
    COUNT(ik.id) as total_keywords
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.is_active = true AND ik.is_active = true;

-- Weight distribution by industry category
SELECT 
    i.category,
    COUNT(ik.id) as total_keywords,
    ROUND(AVG(ik.weight), 3) as avg_weight,
    ROUND(MIN(ik.weight), 3) as min_weight,
    ROUND(MAX(ik.weight), 3) as max_weight,
    COUNT(CASE WHEN ik.weight >= 0.80 THEN 1 END) as high_quality_keywords,
    ROUND(COUNT(CASE WHEN ik.weight >= 0.80 THEN 1 END)::decimal / COUNT(ik.id) * 100, 1) as high_quality_percentage
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.is_active = true
GROUP BY i.category
ORDER BY avg_weight DESC;

-- =============================================================================
-- VALIDATION 4: CROSS-INDUSTRY KEYWORD OVERLAP ANALYSIS
-- =============================================================================
SELECT 
    'VALIDATION 4: CROSS-INDUSTRY KEYWORD OVERLAP ANALYSIS' as validation_name,
    '' as spacer;

-- Check for keyword overlaps between industries
WITH keyword_overlaps AS (
    SELECT 
        ik.keyword,
        COUNT(DISTINCT i.name) as industry_count,
        STRING_AGG(i.name, ', ' ORDER BY i.name) as industries,
        STRING_AGG(i.category, ', ' ORDER BY i.name) as categories
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true AND ik.is_active = true
    GROUP BY ik.keyword
    HAVING COUNT(DISTINCT i.name) > 1
)
SELECT 
    keyword,
    industry_count,
    industries,
    categories,
    CASE 
        WHEN industry_count = 2 THEN 'Low Overlap'
        WHEN industry_count = 3 THEN 'Moderate Overlap'
        WHEN industry_count = 4 THEN 'High Overlap'
        WHEN industry_count >= 5 THEN 'Very High Overlap'
        ELSE 'Unknown'
    END as overlap_level
FROM keyword_overlaps
ORDER BY industry_count DESC, keyword
LIMIT 20;

-- Summary of overlap statistics
SELECT 
    'Overlap Statistics Summary' as analysis_type,
    COUNT(CASE WHEN industry_count = 2 THEN 1 END) as two_industry_overlaps,
    COUNT(CASE WHEN industry_count = 3 THEN 1 END) as three_industry_overlaps,
    COUNT(CASE WHEN industry_count = 4 THEN 1 END) as four_industry_overlaps,
    COUNT(CASE WHEN industry_count >= 5 THEN 1 END) as five_plus_industry_overlaps,
    COUNT(*) as total_overlapping_keywords
FROM (
    SELECT 
        ik.keyword,
        COUNT(DISTINCT i.name) as industry_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true AND ik.is_active = true
    GROUP BY ik.keyword
    HAVING COUNT(DISTINCT i.name) > 1
) overlap_stats;

-- =============================================================================
-- VALIDATION 5: PERFORMANCE AND QUALITY METRICS
-- =============================================================================
SELECT 
    'VALIDATION 5: PERFORMANCE AND QUALITY METRICS' as validation_name,
    '' as spacer;

-- Industries meeting success criteria
SELECT 
    'Industries Meeting Success Criteria' as metric_name,
    COUNT(*) as industries_meeting_criteria,
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM industries WHERE is_active = true) * 100, 1) as percentage
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
    WHERE i.is_active = true
    GROUP BY i.id
    HAVING COUNT(ik.id) >= 20  -- Minimum 20 keywords per industry
) meeting_criteria;

-- High-confidence industries with adequate keywords
SELECT 
    'High-Confidence Industries with Adequate Keywords' as metric_name,
    COUNT(*) as high_confidence_industries,
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM industries WHERE is_active = true AND confidence_threshold >= 0.80) * 100, 1) as percentage
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
    WHERE i.is_active = true AND i.confidence_threshold >= 0.80
    GROUP BY i.id
    HAVING COUNT(ik.id) >= 50  -- High-confidence industries need 50+ keywords
) high_confidence_adequate;

-- Medium-confidence industries with adequate keywords
SELECT 
    'Medium-Confidence Industries with Adequate Keywords' as metric_name,
    COUNT(*) as medium_confidence_industries,
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM industries WHERE is_active = true AND confidence_threshold >= 0.70 AND confidence_threshold < 0.80) * 100, 1) as percentage
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
    WHERE i.is_active = true AND i.confidence_threshold >= 0.70 AND i.confidence_threshold < 0.80
    GROUP BY i.id
    HAVING COUNT(ik.id) >= 30  -- Medium-confidence industries need 30+ keywords
) medium_confidence_adequate;

-- =============================================================================
-- VALIDATION 6: SUCCESS CRITERIA VALIDATION
-- =============================================================================
SELECT 
    'VALIDATION 6: SUCCESS CRITERIA VALIDATION' as validation_name,
    '' as spacer;

-- Validate each success criterion
SELECT 
    'Success Criteria Validation' as validation_type,
    '' as spacer;

-- Criterion 1: 1500+ keywords added across all 39 industries
SELECT 
    'Criterion 1: Total Keywords >= 1500' as criterion,
    COUNT(ik.id) as actual_keywords,
    1500 as target_keywords,
    CASE 
        WHEN COUNT(ik.id) >= 1500 THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.is_active = true AND ik.is_active = true;

-- Criterion 2: Keywords have appropriate base weights (0.5-1.0)
SELECT 
    'Criterion 2: Keywords in Weight Range 0.5-1.0' as criterion,
    COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight <= 1.00 THEN 1 END) as valid_weights,
    COUNT(ik.id) as total_keywords,
    ROUND(COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight <= 1.00 THEN 1 END)::decimal / COUNT(ik.id) * 100, 1) as valid_percentage,
    CASE 
        WHEN COUNT(CASE WHEN ik.weight >= 0.50 AND ik.weight <= 1.00 THEN 1 END) = COUNT(ik.id) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.is_active = true AND ik.is_active = true;

-- Criterion 3: All 39 industries have adequate keyword coverage
SELECT 
    'Criterion 3: All Industries Have Adequate Coverage' as criterion,
    COUNT(CASE WHEN keyword_count >= 20 THEN 1 END) as industries_with_adequate_coverage,
    COUNT(*) as total_industries,
    ROUND(COUNT(CASE WHEN keyword_count >= 20 THEN 1 END)::decimal / COUNT(*) * 100, 1) as coverage_percentage,
    CASE 
        WHEN COUNT(CASE WHEN keyword_count >= 20 THEN 1 END) = COUNT(*) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM (
    SELECT 
        i.id,
        COUNT(ik.id) as keyword_count
    FROM industries i
    LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
    WHERE i.is_active = true
    GROUP BY i.id
) industry_coverage;

-- Criterion 4: No duplicate keywords within industries
SELECT 
    'Criterion 4: No Duplicate Keywords Within Industries' as criterion,
    COUNT(*) as total_keyword_industry_pairs,
    COUNT(DISTINCT CONCAT(ik.keyword, '|', ik.industry_id)) as unique_keyword_industry_pairs,
    CASE 
        WHEN COUNT(*) = COUNT(DISTINCT CONCAT(ik.keyword, '|', ik.industry_id)) THEN 'PASS'
        ELSE 'FAIL'
    END as status
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.is_active = true AND ik.is_active = true;

-- =============================================================================
-- VALIDATION 7: INDUSTRY CATEGORY ANALYSIS
-- =============================================================================
SELECT 
    'VALIDATION 7: INDUSTRY CATEGORY ANALYSIS' as validation_name,
    '' as spacer;

-- Analysis by industry category
SELECT 
    i.category,
    COUNT(DISTINCT i.id) as industry_count,
    COUNT(ik.id) as total_keywords,
    ROUND(COUNT(ik.id)::decimal / COUNT(DISTINCT i.id), 1) as avg_keywords_per_industry,
    ROUND(AVG(ik.weight), 3) as avg_keyword_weight,
    ROUND(AVG(i.confidence_threshold), 3) as avg_confidence_threshold,
    COUNT(CASE WHEN ik.weight >= 0.80 THEN 1 END) as high_quality_keywords,
    ROUND(COUNT(CASE WHEN ik.weight >= 0.80 THEN 1 END)::decimal / COUNT(ik.id) * 100, 1) as high_quality_percentage
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.is_active = true
GROUP BY i.category
ORDER BY avg_keywords_per_industry DESC;

-- =============================================================================
-- VALIDATION 8: FINAL COMPREHENSIVE ASSESSMENT
-- =============================================================================
SELECT 
    'VALIDATION 8: FINAL COMPREHENSIVE ASSESSMENT' as validation_name,
    '' as spacer;

-- Overall assessment
DO $$
DECLARE
    total_industries INTEGER;
    total_keywords INTEGER;
    industries_with_adequate_coverage INTEGER;
    keywords_in_valid_range INTEGER;
    duplicate_keywords INTEGER;
    assessment_score INTEGER := 0;
    max_score INTEGER := 5;
BEGIN
    -- Get total counts
    SELECT COUNT(*) INTO total_industries FROM industries WHERE is_active = true;
    
    SELECT COUNT(*) INTO total_keywords 
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true AND ik.is_active = true;
    
    SELECT COUNT(*) INTO industries_with_adequate_coverage
    FROM (
        SELECT i.id
        FROM industries i
        LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
        WHERE i.is_active = true
        GROUP BY i.id
        HAVING COUNT(ik.id) >= 20
    ) adequate_coverage;
    
    SELECT COUNT(*) INTO keywords_in_valid_range
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true AND ik.is_active = true
    AND ik.weight >= 0.50 AND ik.weight <= 1.00;
    
    SELECT COUNT(*) - COUNT(DISTINCT CONCAT(ik.keyword, '|', ik.industry_id)) INTO duplicate_keywords
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true AND ik.is_active = true;
    
    -- Calculate assessment score
    IF total_keywords >= 1500 THEN assessment_score := assessment_score + 1; END IF;
    IF keywords_in_valid_range = total_keywords THEN assessment_score := assessment_score + 1; END IF;
    IF industries_with_adequate_coverage = total_industries THEN assessment_score := assessment_score + 1; END IF;
    IF duplicate_keywords = 0 THEN assessment_score := assessment_score + 1; END IF;
    IF total_industries >= 35 THEN assessment_score := assessment_score + 1; END IF;
    
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'FINAL COMPREHENSIVE ASSESSMENT';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Total Industries: %', total_industries;
    RAISE NOTICE 'Total Keywords: %', total_keywords;
    RAISE NOTICE 'Industries with Adequate Coverage: %/%', industries_with_adequate_coverage, total_industries;
    RAISE NOTICE 'Keywords in Valid Weight Range: %/%', keywords_in_valid_range, total_keywords;
    RAISE NOTICE 'Duplicate Keywords: %', duplicate_keywords;
    RAISE NOTICE 'Assessment Score: %/%', assessment_score, max_score;
    
    IF assessment_score = max_score THEN
        RAISE NOTICE '✅ EXCELLENT: All success criteria met!';
        RAISE NOTICE 'Task 3.2: Comprehensive Keyword Sets - COMPLETED SUCCESSFULLY';
    ELSIF assessment_score >= 4 THEN
        RAISE NOTICE '✅ GOOD: Most success criteria met';
        RAISE NOTICE 'Task 3.2: Comprehensive Keyword Sets - MOSTLY COMPLETED';
    ELSIF assessment_score >= 3 THEN
        RAISE NOTICE '⚠️ FAIR: Some success criteria met';
        RAISE NOTICE 'Task 3.2: Comprehensive Keyword Sets - PARTIALLY COMPLETED';
    ELSE
        RAISE NOTICE '❌ POOR: Few success criteria met';
        RAISE NOTICE 'Task 3.2: Comprehensive Keyword Sets - NEEDS IMPROVEMENT';
    END IF;
    
    RAISE NOTICE '=============================================================================';
END $$;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'COMPREHENSIVE INDUSTRY KEYWORD VALIDATION COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Validation Categories: 8 comprehensive validation suites';
    RAISE NOTICE 'Analysis: Overall coverage, weight distribution, overlap, performance';
    RAISE NOTICE 'Assessment: Success criteria validation and final scoring';
    RAISE NOTICE 'Status: Ready for Phase 4 - Testing & Validation';
    RAISE NOTICE 'Next: Proceed to Task 4.1 - Comprehensive Test Suite';
    RAISE NOTICE '=============================================================================';
END $$;
