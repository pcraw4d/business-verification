-- =============================================================================
-- TASK 3.2.1: VALIDATE ALL INDUSTRIES KEYWORD COVERAGE
-- =============================================================================
-- This script validates that all industries in the database have adequate
-- keyword coverage for accurate classification (>85% accuracy target).
-- =============================================================================

-- =============================================================================
-- VALIDATION 1: OVERALL INDUSTRY COVERAGE
-- =============================================================================
SELECT 
    'VALIDATION 1: OVERALL INDUSTRY COVERAGE' as validation_name,
    '' as spacer;

-- Check total industries and their keyword coverage
SELECT 
    COUNT(DISTINCT i.id) as total_industries,
    COUNT(DISTINCT CASE WHEN i.is_active = true THEN i.id END) as active_industries,
    COUNT(DISTINCT CASE WHEN kw.is_active = true THEN i.id END) as industries_with_keywords,
    COUNT(kw.id) as total_keywords,
    COUNT(CASE WHEN kw.is_active = true THEN 1 END) as active_keywords,
    ROUND(COUNT(kw.id)::decimal / COUNT(DISTINCT i.id), 1) as avg_keywords_per_industry
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id
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
    COUNT(kw.id) as keyword_count,
    COUNT(CASE WHEN kw.base_weight >= 0.90 THEN 1 END) as excellent_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END) as high_quality_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.70 THEN 1 END) as good_keywords,
    ROUND(AVG(kw.base_weight), 3) as avg_keyword_weight,
    CASE 
        WHEN COUNT(kw.id) >= 50 THEN 'Excellent'
        WHEN COUNT(kw.id) >= 30 THEN 'Good'
        WHEN COUNT(kw.id) >= 20 THEN 'Fair'
        WHEN COUNT(kw.id) >= 10 THEN 'Poor'
        ELSE 'Inadequate'
    END as coverage_quality,
    CASE 
        WHEN COUNT(kw.id) >= 20 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_result
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.category, i.confidence_threshold
ORDER BY 
    CASE 
        WHEN COUNT(kw.id) >= 50 THEN 1
        WHEN COUNT(kw.id) >= 30 THEN 2
        WHEN COUNT(kw.id) >= 20 THEN 3
        WHEN COUNT(kw.id) >= 10 THEN 4
        ELSE 5
    END,
    i.name;

-- =============================================================================
-- VALIDATION 3: KEYWORD QUALITY DISTRIBUTION
-- =============================================================================
SELECT 
    'VALIDATION 3: KEYWORD QUALITY DISTRIBUTION' as validation_name,
    '' as spacer;

-- Check keyword quality distribution across all industries
SELECT 
    'All Industries' as scope,
    COUNT(kw.id) as total_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.95 THEN 1 END) as excellent_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.90 AND kw.base_weight < 0.95 THEN 1 END) as very_good_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.80 AND kw.base_weight < 0.90 THEN 1 END) as good_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.70 AND kw.base_weight < 0.80 THEN 1 END) as fair_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.50 AND kw.base_weight < 0.70 THEN 1 END) as poor_keywords,
    ROUND(AVG(kw.base_weight), 3) as average_weight,
    ROUND(COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END)::decimal / COUNT(kw.id) * 100, 1) as high_quality_percentage
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true;

-- =============================================================================
-- VALIDATION 4: CATEGORY-BASED COVERAGE
-- =============================================================================
SELECT 
    'VALIDATION 4: CATEGORY-BASED COVERAGE' as validation_name,
    '' as spacer;

-- Check keyword coverage by industry category
SELECT 
    i.category,
    COUNT(DISTINCT i.id) as industries_in_category,
    COUNT(kw.id) as total_keywords,
    ROUND(COUNT(kw.id)::decimal / COUNT(DISTINCT i.id), 1) as avg_keywords_per_industry,
    COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END) as high_quality_keywords,
    ROUND(COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END)::decimal / COUNT(kw.id) * 100, 1) as high_quality_percentage,
    CASE 
        WHEN COUNT(kw.id)::decimal / COUNT(DISTINCT i.id) >= 30 THEN 'Excellent'
        WHEN COUNT(kw.id)::decimal / COUNT(DISTINCT i.id) >= 20 THEN 'Good'
        WHEN COUNT(kw.id)::decimal / COUNT(DISTINCT i.id) >= 10 THEN 'Fair'
        ELSE 'Poor'
    END as category_coverage_quality
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.is_active = true
GROUP BY i.category
ORDER BY avg_keywords_per_industry DESC;

-- =============================================================================
-- VALIDATION 5: INDUSTRIES NEEDING KEYWORD EXPANSION
-- =============================================================================
SELECT 
    'VALIDATION 5: INDUSTRIES NEEDING KEYWORD EXPANSION' as validation_name,
    '' as spacer;

-- Identify industries that need more keywords
SELECT 
    i.name as industry_name,
    i.category,
    i.confidence_threshold,
    COUNT(kw.id) as current_keywords,
    CASE 
        WHEN i.confidence_threshold >= 0.80 THEN 50
        WHEN i.confidence_threshold >= 0.75 THEN 40
        WHEN i.confidence_threshold >= 0.70 THEN 30
        ELSE 20
    END as recommended_keywords,
    CASE 
        WHEN i.confidence_threshold >= 0.80 THEN 50
        WHEN i.confidence_threshold >= 0.75 THEN 40
        WHEN i.confidence_threshold >= 0.70 THEN 30
        ELSE 20
    END - COUNT(kw.id) as keywords_needed,
    CASE 
        WHEN COUNT(kw.id) < 20 THEN 'CRITICAL - Needs immediate attention'
        WHEN COUNT(kw.id) < 30 THEN 'HIGH - Needs expansion'
        WHEN COUNT(kw.id) < 40 THEN 'MEDIUM - Could use more keywords'
        ELSE 'GOOD - Adequate coverage'
    END as priority_level
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.category, i.confidence_threshold
HAVING COUNT(kw.id) < 50
ORDER BY 
    CASE 
        WHEN COUNT(kw.id) < 20 THEN 1
        WHEN COUNT(kw.id) < 30 THEN 2
        WHEN COUNT(kw.id) < 40 THEN 3
        ELSE 4
    END,
    keywords_needed DESC;

-- =============================================================================
-- VALIDATION 6: KEYWORD WEIGHT DISTRIBUTION
-- =============================================================================
SELECT 
    'VALIDATION 6: KEYWORD WEIGHT DISTRIBUTION' as validation_name,
    '' as spacer;

-- Check that keyword weights are properly distributed
SELECT 
    'Weight Range' as metric,
    COUNT(*) as keyword_count,
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) * 100, 1) as percentage
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true
AND kw.base_weight >= 0.90
UNION ALL
SELECT 
    '0.80-0.89',
    COUNT(*),
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) * 100, 1)
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true
AND kw.base_weight >= 0.80 AND kw.base_weight < 0.90
UNION ALL
SELECT 
    '0.70-0.79',
    COUNT(*),
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) * 100, 1)
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true
AND kw.base_weight >= 0.70 AND kw.base_weight < 0.80
UNION ALL
SELECT 
    '0.50-0.69',
    COUNT(*),
    ROUND(COUNT(*)::decimal / (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) * 100, 1)
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true
AND kw.base_weight >= 0.50 AND kw.base_weight < 0.70
ORDER BY 
    CASE 
        WHEN metric = 'Weight Range' THEN 1
        WHEN metric = '0.80-0.89' THEN 2
        WHEN metric = '0.70-0.79' THEN 3
        WHEN metric = '0.50-0.69' THEN 4
    END;

-- =============================================================================
-- VALIDATION 7: CONFIDENCE THRESHOLD ALIGNMENT
-- =============================================================================
SELECT 
    'VALIDATION 7: CONFIDENCE THRESHOLD ALIGNMENT' as validation_name,
    '' as spacer;

-- Check that industries have enough high-quality keywords to meet their confidence thresholds
SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(kw.id) as total_keywords,
    COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END) as keywords_above_threshold,
    ROUND(COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END)::decimal / COUNT(kw.id) * 100, 1) as percentage_above_threshold,
    CASE 
        WHEN COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END) >= 15 THEN 'Excellent'
        WHEN COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END) >= 10 THEN 'Good'
        WHEN COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END) >= 5 THEN 'Fair'
        ELSE 'Poor'
    END as threshold_alignment,
    CASE 
        WHEN COUNT(CASE WHEN kw.base_weight >= i.confidence_threshold THEN 1 END) >= 10 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_result
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.confidence_threshold DESC;

-- =============================================================================
-- VALIDATION 8: DUPLICATE KEYWORD CHECK
-- =============================================================================
SELECT 
    'VALIDATION 8: DUPLICATE KEYWORD CHECK' as validation_name,
    '' as spacer;

-- Check for duplicate keywords across industries
SELECT 
    kw.keyword,
    COUNT(*) as duplicate_count,
    STRING_AGG(i.name, ', ') as industries,
    CASE 
        WHEN COUNT(*) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as validation_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = true
GROUP BY kw.keyword
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- =============================================================================
-- VALIDATION 9: CLASSIFICATION READINESS ASSESSMENT
-- =============================================================================
SELECT 
    'VALIDATION 9: CLASSIFICATION READINESS ASSESSMENT' as validation_name,
    '' as spacer;

-- Assess overall readiness for >85% classification accuracy
SELECT 
    'Overall System' as assessment_scope,
    COUNT(DISTINCT i.id) as total_industries,
    COUNT(kw.id) as total_keywords,
    COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END) as high_quality_keywords,
    ROUND(COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END)::decimal / COUNT(kw.id) * 100, 1) as high_quality_percentage,
    COUNT(CASE WHEN COUNT(kw2.id) >= 20 THEN 1 END) as industries_with_adequate_keywords,
    ROUND(COUNT(CASE WHEN COUNT(kw2.id) >= 20 THEN 1 END)::decimal / COUNT(DISTINCT i.id) * 100, 1) as industries_ready_percentage,
    CASE 
        WHEN COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END) >= 100 
         AND COUNT(CASE WHEN COUNT(kw2.id) >= 20 THEN 1 END) >= COUNT(DISTINCT i.id) * 0.8 THEN 'READY FOR >85% ACCURACY'
        WHEN COUNT(CASE WHEN kw.base_weight >= 0.80 THEN 1 END) >= 50 
         AND COUNT(CASE WHEN COUNT(kw2.id) >= 20 THEN 1 END) >= COUNT(DISTINCT i.id) * 0.6 THEN 'READY FOR >80% ACCURACY'
        ELSE 'NEEDS IMPROVEMENT'
    END as readiness_status
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
LEFT JOIN keyword_weights kw2 ON i.id = kw2.industry_id AND kw2.is_active = true
WHERE i.is_active = true
GROUP BY i.id;

-- =============================================================================
-- VALIDATION 10: FINAL RECOMMENDATIONS
-- =============================================================================
SELECT 
    'VALIDATION 10: FINAL RECOMMENDATIONS' as validation_name,
    '' as spacer;

-- Provide final recommendations for achieving >85% accuracy
SELECT 
    'RECOMMENDATIONS FOR >85% CLASSIFICATION ACCURACY' as recommendation_type,
    '' as spacer;

-- Summary of what needs to be done
SELECT 
    'Priority 1: Critical Industries' as priority,
    COUNT(*) as industries_needing_attention,
    'Industries with <20 keywords need immediate keyword expansion' as action_required
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
    WHERE i.is_active = true
    GROUP BY i.id
    HAVING COUNT(kw.id) < 20
) critical_industries

UNION ALL

SELECT 
    'Priority 2: High-Confidence Industries',
    COUNT(*),
    'Industries with confidence_threshold >= 0.80 need >= 50 high-quality keywords'
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
    WHERE i.is_active = true AND i.confidence_threshold >= 0.80
    GROUP BY i.id
    HAVING COUNT(kw.id) < 50
) high_confidence_industries

UNION ALL

SELECT 
    'Priority 3: Medium-Confidence Industries',
    COUNT(*),
    'Industries with confidence_threshold >= 0.70 need >= 30 keywords'
FROM (
    SELECT i.id
    FROM industries i
    LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
    WHERE i.is_active = true AND i.confidence_threshold >= 0.70 AND i.confidence_threshold < 0.80
    GROUP BY i.id
    HAVING COUNT(kw.id) < 30
) medium_confidence_industries;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TASK 3.2.1: ALL INDUSTRIES KEYWORD COVERAGE VALIDATION COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Validations performed: 10 comprehensive coverage assessments';
    RAISE NOTICE 'Industries analyzed: All active industries in database';
    RAISE NOTICE 'Keywords validated: All active keywords across all industries';
    RAISE NOTICE 'Quality assessed: Keyword weight distribution and quality';
    RAISE NOTICE 'Readiness evaluated: Classification accuracy readiness';
    RAISE NOTICE 'Recommendations provided: Priority-based keyword expansion plan';
    RAISE NOTICE 'Status: Ready for next phase of keyword expansion';
    RAISE NOTICE 'Next: Implement keyword expansion for industries needing attention';
    RAISE NOTICE '=============================================================================';
END $$;
