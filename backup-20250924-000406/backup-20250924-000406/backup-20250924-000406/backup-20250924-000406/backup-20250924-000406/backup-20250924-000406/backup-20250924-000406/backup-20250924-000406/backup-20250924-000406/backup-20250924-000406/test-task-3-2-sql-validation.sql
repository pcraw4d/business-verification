-- Task 3.2: Comprehensive Keyword Sets Testing
-- This script executes all the testing procedures outlined in the comprehensive plan
-- for validating the keyword sets across all 39 industries

-- =====================================================
-- Test 1: Verify keyword count per industry
-- Expected: Each industry should have 20+ keywords
-- =====================================================

SELECT 
    'Test 1: Keyword Count Per Industry' as test_name,
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    CASE 
        WHEN COUNT(kw.keyword) >= 20 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industries i
LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
WHERE i.is_active = true
GROUP BY i.id, i.name
ORDER BY keyword_count DESC;

-- Summary for Test 1
SELECT 
    'Test 1 Summary' as summary,
    COUNT(*) as total_industries,
    COUNT(CASE WHEN keyword_count >= 20 THEN 1 END) as industries_with_adequate_keywords,
    ROUND(COUNT(CASE WHEN keyword_count >= 20 THEN 1 END) * 100.0 / COUNT(*), 2) as coverage_percentage
FROM (
    SELECT i.name, COUNT(kw.keyword) as keyword_count
    FROM industries i
    LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
    WHERE i.is_active = true
    GROUP BY i.id, i.name
) industry_counts;

-- =====================================================
-- Test 2: Verify keyword weights distribution
-- Expected: All keywords should have base weights between 0.5 and 1.0
-- =====================================================

SELECT 
    'Test 2: Keyword Weights Distribution' as test_name,
    i.name as industry_name,
    MIN(kw.base_weight) as min_weight,
    MAX(kw.base_weight) as max_weight,
    ROUND(AVG(kw.base_weight), 4) as avg_weight,
    COUNT(kw.keyword) as total_keywords,
    CASE 
        WHEN MIN(kw.base_weight) >= 0.5 AND MAX(kw.base_weight) <= 1.0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.is_active = true AND kw.is_active = true
GROUP BY i.id, i.name
ORDER BY avg_weight DESC;

-- Summary for Test 2
SELECT 
    'Test 2 Summary' as summary,
    COUNT(*) as total_keywords,
    MIN(base_weight) as global_min_weight,
    MAX(base_weight) as global_max_weight,
    ROUND(AVG(base_weight), 4) as global_avg_weight,
    COUNT(CASE WHEN base_weight >= 0.5 AND base_weight <= 1.0 THEN 1 END) as keywords_in_range,
    ROUND(COUNT(CASE WHEN base_weight >= 0.5 AND base_weight <= 1.0 THEN 1 END) * 100.0 / COUNT(*), 2) as range_coverage_percentage
FROM keyword_weights 
WHERE is_active = true;

-- =====================================================
-- Test 3: Verify no duplicate keywords within industries
-- Expected: No duplicate keywords within the same industry
-- =====================================================

SELECT 
    'Test 3: Duplicate Keywords Within Industries' as test_name,
    i.name as industry_name,
    kw.keyword,
    COUNT(*) as duplicate_count,
    CASE 
        WHEN COUNT(*) = 1 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.is_active = true AND kw.is_active = true
GROUP BY i.id, i.name, kw.keyword
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, i.name, kw.keyword;

-- Summary for Test 3
SELECT 
    'Test 3 Summary' as summary,
    COUNT(*) as duplicate_keyword_instances,
    COUNT(DISTINCT i.name) as industries_with_duplicates,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS - No duplicates found'
        ELSE 'FAIL - Duplicates found'
    END as overall_result
FROM (
    SELECT i.name, kw.keyword, COUNT(*) as duplicate_count
    FROM industries i
    JOIN keyword_weights kw ON i.id = kw.industry_id
    WHERE i.is_active = true AND kw.is_active = true
    GROUP BY i.id, i.name, kw.keyword
    HAVING COUNT(*) > 1
) duplicates;

-- =====================================================
-- Test 4: Verify industry coverage for >85% accuracy
-- Expected: All 39 industries should have adequate coverage
-- =====================================================

-- Industry coverage by category
SELECT 
    'Test 4: Industry Coverage by Category' as test_name,
    i.category,
    COUNT(*) as industries_in_category,
    SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) as industries_with_adequate_keywords,
    ROUND(SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as coverage_percentage,
    CASE 
        WHEN SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) = COUNT(*) THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM industries i
LEFT JOIN (
    SELECT industry_id, COUNT(keyword) as keyword_count
    FROM keyword_weights 
    WHERE is_active = true
    GROUP BY industry_id
) kw_count ON i.id = kw_count.industry_id
WHERE i.is_active = true
GROUP BY i.category
ORDER BY coverage_percentage DESC;

-- Overall coverage summary
SELECT 
    'Test 4 Summary' as summary,
    COUNT(*) as total_industries,
    SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) as industries_with_adequate_coverage,
    ROUND(SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as overall_coverage_percentage,
    CASE 
        WHEN SUM(CASE WHEN kw_count.keyword_count >= 20 THEN 1 ELSE 0 END) = COUNT(*) THEN 'PASS - All industries have adequate coverage'
        ELSE 'FAIL - Some industries lack adequate coverage'
    END as overall_result
FROM industries i
LEFT JOIN (
    SELECT industry_id, COUNT(keyword) as keyword_count
    FROM keyword_weights 
    WHERE is_active = true
    GROUP BY industry_id
) kw_count ON i.id = kw_count.industry_id
WHERE i.is_active = true;

-- =====================================================
-- Test 5: Verify all 7 subtasks completed successfully
-- Expected: All subtasks should have comprehensive keyword sets
-- =====================================================

-- Legal Services (3.2.1)
SELECT 
    'Test 5.1: Legal Services Keywords (3.2.1)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property')
AND kw.is_active = true AND i.is_active = true;

-- Healthcare (3.2.2)
SELECT 
    'Test 5.2: Healthcare Keywords (3.2.2)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology')
AND kw.is_active = true AND i.is_active = true;

-- Technology (3.2.3)
SELECT 
    'Test 5.3: Technology Keywords (3.2.3)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence', 
                 'Technology Services', 'Digital Services', 'EdTech', 'Industrial Technology', 
                 'Food Technology', 'Healthcare Technology', 'Fintech')
AND kw.is_active = true AND i.is_active = true;

-- Retail & E-commerce (3.2.4)
SELECT 
    'Test 5.4: Retail & E-commerce Keywords (3.2.4)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Retail', 'E-commerce', 'Wholesale', 'Consumer Goods')
AND kw.is_active = true AND i.is_active = true;

-- Manufacturing (3.2.5)
SELECT 
    'Test 5.5: Manufacturing Keywords (3.2.5)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing')
AND kw.is_active = true AND i.is_active = true;

-- Financial Services (3.2.6)
SELECT 
    'Test 5.6: Financial Services Keywords (3.2.6)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Banking', 'Insurance', 'Investment Services')
AND kw.is_active = true AND i.is_active = true;

-- Agriculture & Energy (3.2.7)
SELECT 
    'Test 5.7: Agriculture & Energy Keywords (3.2.7)' as test_name,
    COUNT(*) as keyword_count,
    CASE 
        WHEN COUNT(*) >= 200 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.name IN ('Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy')
AND kw.is_active = true AND i.is_active = true;

-- =====================================================
-- Test 6: Performance and Data Integrity Validation
-- =====================================================

-- Check for orphaned keyword_weights (no corresponding industry)
SELECT 
    'Test 6.1: Orphaned Keywords' as test_name,
    COUNT(*) as orphaned_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
LEFT JOIN industries i ON kw.industry_id = i.id
WHERE i.id IS NULL;

-- Check for inactive keywords in active industries
SELECT 
    'Test 6.2: Inactive Keywords in Active Industries' as test_name,
    COUNT(*) as inactive_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights kw
JOIN industries i ON kw.industry_id = i.id
WHERE i.is_active = true AND kw.is_active = false;

-- Check for null or empty keywords
SELECT 
    'Test 6.3: Null or Empty Keywords' as test_name,
    COUNT(*) as invalid_count,
    CASE 
        WHEN COUNT(*) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as test_result
FROM keyword_weights 
WHERE keyword IS NULL OR TRIM(keyword) = '';

-- =====================================================
-- Final Summary Report
-- =====================================================

SELECT 
    'FINAL SUMMARY' as report_section,
    'Task 3.2 Comprehensive Keyword Sets Testing' as test_suite,
    NOW() as test_timestamp;

-- Overall statistics
SELECT 
    'Overall Statistics' as metric,
    (SELECT COUNT(*) FROM industries WHERE is_active = true) as total_industries,
    (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) as total_keywords,
    (SELECT COUNT(DISTINCT industry_id) FROM keyword_weights WHERE is_active = true) as industries_with_keywords,
    (SELECT ROUND(AVG(keyword_count), 2) FROM (
        SELECT COUNT(kw.keyword) as keyword_count
        FROM industries i
        LEFT JOIN keyword_weights kw ON i.id = kw.industry_id AND kw.is_active = true
        WHERE i.is_active = true
        GROUP BY i.id
    ) counts) as avg_keywords_per_industry;

-- Success criteria validation
SELECT 
    'Success Criteria Validation' as criteria,
    CASE 
        WHEN (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) >= 1500 THEN 'PASS'
        ELSE 'FAIL'
    END as keyword_count_target,
    CASE 
        WHEN (SELECT COUNT(*) FROM industries WHERE is_active = true) >= 39 THEN 'PASS'
        ELSE 'FAIL'
    END as industry_count_target,
    CASE 
        WHEN (SELECT COUNT(*) FROM keyword_weights WHERE base_weight >= 0.5 AND base_weight <= 1.0 AND is_active = true) = 
             (SELECT COUNT(*) FROM keyword_weights WHERE is_active = true) THEN 'PASS'
        ELSE 'FAIL'
    END as weight_range_target,
    CASE 
        WHEN (SELECT COUNT(*) FROM (
            SELECT i.name, kw.keyword, COUNT(*) as duplicate_count
            FROM industries i
            JOIN keyword_weights kw ON i.id = kw.industry_id
            WHERE i.is_active = true AND kw.is_active = true
            GROUP BY i.id, i.name, kw.keyword
            HAVING COUNT(*) > 1
        ) duplicates) = 0 THEN 'PASS'
        ELSE 'FAIL'
    END as no_duplicates_target;
