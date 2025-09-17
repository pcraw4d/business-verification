-- =============================================================================
-- TASK 5.1.2: TEST EXPANDED KEYWORD COVERAGE AND RELATIONSHIP MAPPING
-- =============================================================================
-- This script tests the keyword expansion functionality, relationship mapping,
-- and validates that the 2000+ keyword expansion achieves target accuracy.
-- =============================================================================

-- Start transaction for testing
BEGIN;

-- =============================================================================
-- 1. VALIDATE KEYWORD RELATIONSHIP TABLES
-- =============================================================================

-- Test 1: Verify keyword relationships table exists and has data
SELECT 
    'TEST 1: KEYWORD RELATIONSHIPS TABLE' as test_name,
    COUNT(*) as total_relationships,
    COUNT(CASE WHEN relationship_type = 'synonym' THEN 1 END) as synonyms,
    COUNT(CASE WHEN relationship_type = 'abbreviation' THEN 1 END) as abbreviations,
    COUNT(CASE WHEN relationship_type = 'related' THEN 1 END) as related_terms,
    COUNT(CASE WHEN relationship_type = 'variant' THEN 1 END) as variants,
    AVG(confidence_score) as avg_confidence_score,
    MIN(confidence_score) as min_confidence,
    MAX(confidence_score) as max_confidence
FROM keyword_relationships
WHERE is_active = true;

-- Test 2: Verify keyword contexts table exists and has data
SELECT 
    'TEST 2: KEYWORD CONTEXTS TABLE' as test_name,
    COUNT(*) as total_contexts,
    COUNT(CASE WHEN context_type = 'primary' THEN 1 END) as primary_contexts,
    COUNT(CASE WHEN context_type = 'technical' THEN 1 END) as technical_contexts,
    COUNT(CASE WHEN context_type = 'business' THEN 1 END) as business_contexts,
    COUNT(CASE WHEN context_type = 'secondary' THEN 1 END) as secondary_contexts,
    COUNT(CASE WHEN context_type = 'general' THEN 1 END) as general_contexts,
    AVG(context_weight) as avg_context_weight,
    MIN(context_weight) as min_weight,
    MAX(context_weight) as max_weight
FROM keyword_contexts
WHERE is_active = true;

-- =============================================================================
-- 2. TEST KEYWORD EXPANSION SCENARIOS
-- =============================================================================

-- Test 3: Technology keyword expansion
SELECT 
    'TEST 3: TECHNOLOGY KEYWORD EXPANSION' as test_name,
    '' as spacer;

-- Test expansion for 'software'
SELECT 
    'SOFTWARE EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'software' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- Test expansion for 'technology'
SELECT 
    'TECHNOLOGY EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'technology' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- Test 4: Healthcare keyword expansion
SELECT 
    'TEST 4: HEALTHCARE KEYWORD EXPANSION' as test_name,
    '' as spacer;

-- Test expansion for 'medical'
SELECT 
    'MEDICAL EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'medical' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- Test expansion for 'healthcare'
SELECT 
    'HEALTHCARE EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'healthcare' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- Test 5: Financial Services keyword expansion
SELECT 
    'TEST 5: FINANCIAL SERVICES KEYWORD EXPANSION' as test_name,
    '' as spacer;

-- Test expansion for 'banking'
SELECT 
    'BANKING EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'banking' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- Test expansion for 'investment'
SELECT 
    'INVESTMENT EXPANSION' as keyword,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score
FROM keyword_relationships kr
WHERE kr.primary_keyword = 'investment' 
  AND kr.is_active = true
ORDER BY kr.confidence_score DESC;

-- =============================================================================
-- 3. TEST KEYWORD CONTEXTS
-- =============================================================================

-- Test 6: Context mapping for technology industry
SELECT 
    'TEST 6: TECHNOLOGY INDUSTRY CONTEXTS' as test_name,
    '' as spacer;

SELECT 
    kc.keyword,
    kc.context_type,
    kc.context_weight,
    i.name as industry_name
FROM keyword_contexts kc
JOIN industries i ON kc.industry_id = i.id
WHERE i.name = 'Technology' 
  AND kc.is_active = true
ORDER BY kc.context_weight DESC, kc.keyword
LIMIT 20;

-- Test 7: Context mapping for healthcare industry
SELECT 
    'TEST 7: HEALTHCARE INDUSTRY CONTEXTS' as test_name,
    '' as spacer;

SELECT 
    kc.keyword,
    kc.context_type,
    kc.context_weight,
    i.name as industry_name
FROM keyword_contexts kc
JOIN industries i ON kc.industry_id = i.id
WHERE i.name = 'Healthcare' 
  AND kc.is_active = true
ORDER BY kc.context_weight DESC, kc.keyword
LIMIT 20;

-- =============================================================================
-- 4. VALIDATE KEYWORD COVERAGE EXPANSION
-- =============================================================================

-- Test 8: Total keyword count validation
SELECT 
    'TEST 8: TOTAL KEYWORD COUNT VALIDATION' as test_name,
    '' as spacer;

-- Count keywords by industry
SELECT 
    i.name as industry_name,
    COUNT(ik.id) as keyword_count,
    COUNT(CASE WHEN ik.is_primary = true THEN 1 END) as primary_keywords,
    COUNT(CASE WHEN ik.context = 'technical' THEN 1 END) as technical_keywords,
    COUNT(CASE WHEN ik.context = 'business' THEN 1 END) as business_keywords,
    AVG(ik.weight) as avg_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.is_active = true
GROUP BY i.id, i.name
ORDER BY keyword_count DESC;

-- Test 9: Overall keyword expansion summary
SELECT 
    'TEST 9: KEYWORD EXPANSION SUMMARY' as test_name,
    COUNT(DISTINCT ik.keyword) as total_unique_keywords,
    COUNT(ik.id) as total_keyword_instances,
    COUNT(DISTINCT i.id) as industries_with_keywords,
    AVG(keyword_count.count) as avg_keywords_per_industry,
    MIN(keyword_count.count) as min_keywords_per_industry,
    MAX(keyword_count.count) as max_keywords_per_industry
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
JOIN (
    SELECT industry_id, COUNT(*) as count
    FROM industry_keywords
    GROUP BY industry_id
) keyword_count ON i.id = keyword_count.industry_id
WHERE i.is_active = true;

-- =============================================================================
-- 5. TEST RELATIONSHIP MAPPING QUERIES
-- =============================================================================

-- Test 10: Cross-industry keyword relationships
SELECT 
    'TEST 10: CROSS-INDUSTRY KEYWORD RELATIONSHIPS' as test_name,
    '' as spacer;

-- Find keywords that appear in multiple industries with relationships
SELECT 
    kr.primary_keyword,
    COUNT(DISTINCT ik.industry_id) as industry_count,
    STRING_AGG(DISTINCT i.name, ', ') as industries,
    COUNT(DISTINCT kr.related_keyword) as related_keyword_count,
    AVG(kr.confidence_score) as avg_confidence
FROM keyword_relationships kr
JOIN industry_keywords ik ON kr.primary_keyword = ik.keyword
JOIN industries i ON ik.industry_id = i.id
WHERE kr.is_active = true 
  AND ik.is_active = true 
  AND i.is_active = true
GROUP BY kr.primary_keyword
HAVING COUNT(DISTINCT ik.industry_id) > 1
ORDER BY industry_count DESC, avg_confidence DESC
LIMIT 15;

-- Test 11: High-confidence relationship mappings
SELECT 
    'TEST 11: HIGH-CONFIDENCE RELATIONSHIPS' as test_name,
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score,
    CASE 
        WHEN kr.confidence_score >= 0.90 THEN 'Excellent'
        WHEN kr.confidence_score >= 0.80 THEN 'Good'
        WHEN kr.confidence_score >= 0.70 THEN 'Fair'
        ELSE 'Low'
    END as confidence_level
FROM keyword_relationships kr
WHERE kr.is_active = true 
  AND kr.confidence_score >= 0.80
ORDER BY kr.confidence_score DESC, kr.primary_keyword
LIMIT 25;

-- =============================================================================
-- 6. VALIDATE SEARCH PERFORMANCE
-- =============================================================================

-- Test 12: Keyword lookup performance test
SELECT 
    'TEST 12: KEYWORD LOOKUP PERFORMANCE' as test_name,
    '' as spacer;

-- Test query performance for keyword expansion
EXPLAIN (ANALYZE, BUFFERS)
SELECT 
    kr.primary_keyword,
    kr.related_keyword,
    kr.relationship_type,
    kr.confidence_score,
    kc.context_weight
FROM keyword_relationships kr
LEFT JOIN keyword_contexts kc ON kr.related_keyword = kc.keyword
WHERE kr.primary_keyword IN ('technology', 'software', 'healthcare', 'medical', 'banking', 'finance')
  AND kr.is_active = true
  AND kr.confidence_score >= 0.70
ORDER BY kr.confidence_score DESC;

-- =============================================================================
-- 7. ACCURACY VALIDATION TESTS
-- =============================================================================

-- Test 13: Sample business classification tests
SELECT 
    'TEST 13: SAMPLE BUSINESS CLASSIFICATION' as test_name,
    '' as spacer;

-- Test sample technology businesses (simulated)
WITH test_businesses AS (
    SELECT 'TechCorp Software Solutions' as business_name, 'software development ai machine learning' as keywords
    UNION ALL
    SELECT 'HealthTech Medical Systems', 'healthcare medical technology telemedicine'
    UNION ALL
    SELECT 'FinanceFirst Banking Solutions', 'banking financial services investment'
    UNION ALL
    SELECT 'RetailPro E-commerce Platform', 'retail ecommerce online shopping'
    UNION ALL
    SELECT 'LegalTech Document Automation', 'legal technology document automation'
),
keyword_matches AS (
    SELECT 
        tb.business_name,
        tb.keywords,
        STRING_TO_ARRAY(tb.keywords, ' ') as keyword_array
    FROM test_businesses tb
)
SELECT 
    km.business_name,
    km.keywords as test_keywords,
    COUNT(DISTINCT ik.industry_id) as matching_industries,
    STRING_AGG(DISTINCT i.name, ', ') as potential_industries
FROM keyword_matches km
CROSS JOIN UNNEST(km.keyword_array) as keyword
JOIN industry_keywords ik ON LOWER(ik.keyword) = LOWER(keyword)
JOIN industries i ON ik.industry_id = i.id
WHERE ik.is_active = true AND i.is_active = true
GROUP BY km.business_name, km.keywords
ORDER BY matching_industries DESC;

-- Test 14: Keyword relationship effectiveness
SELECT 
    'TEST 14: RELATIONSHIP EFFECTIVENESS' as test_name,
    COUNT(*) as total_primary_keywords,
    COUNT(CASE WHEN relationship_count > 0 THEN 1 END) as keywords_with_relationships,
    ROUND(
        COUNT(CASE WHEN relationship_count > 0 THEN 1 END)::decimal / COUNT(*)::decimal * 100, 
        2
    ) as coverage_percentage,
    AVG(relationship_count) as avg_relationships_per_keyword,
    MAX(relationship_count) as max_relationships
FROM (
    SELECT 
        ik.keyword,
        COUNT(kr.id) as relationship_count
    FROM industry_keywords ik
    LEFT JOIN keyword_relationships kr ON ik.keyword = kr.primary_keyword AND kr.is_active = true
    WHERE ik.is_active = true
    GROUP BY ik.keyword
) keyword_stats;

-- =============================================================================
-- 8. VALIDATION SUMMARY
-- =============================================================================

-- Test 15: Overall validation summary
SELECT 
    'TEST 15: OVERALL VALIDATION SUMMARY' as test_name,
    '' as spacer;

-- Summary statistics
SELECT 
    'KEYWORD EXPANSION SUMMARY' as metric_category,
    COUNT(DISTINCT ik.keyword) as total_keywords,
    COUNT(DISTINCT kr.primary_keyword) as keywords_with_relationships,
    COUNT(DISTINCT kr.related_keyword) as related_keywords,
    COUNT(DISTINCT kc.keyword) as keywords_with_context,
    COUNT(DISTINCT i.id) as active_industries,
    ROUND(AVG(ik.weight), 3) as avg_keyword_weight,
    ROUND(AVG(kr.confidence_score), 3) as avg_relationship_confidence,
    ROUND(AVG(kc.context_weight), 3) as avg_context_weight
FROM industry_keywords ik
FULL OUTER JOIN keyword_relationships kr ON ik.keyword = kr.primary_keyword AND kr.is_active = true
FULL OUTER JOIN keyword_contexts kc ON ik.keyword = kc.keyword AND kc.is_active = true
JOIN industries i ON ik.industry_id = i.id
WHERE ik.is_active = true AND i.is_active = true;

-- Final validation: Check if we've achieved the 2000+ keyword target
SELECT 
    'FINAL VALIDATION: 2000+ KEYWORD TARGET' as validation,
    COUNT(DISTINCT keyword) as unique_keywords,
    CASE 
        WHEN COUNT(DISTINCT keyword) >= 2000 THEN '✅ TARGET ACHIEVED'
        ELSE '❌ TARGET NOT ACHIEVED'
    END as target_status,
    (2000 - COUNT(DISTINCT keyword)) as keywords_needed
FROM industry_keywords
WHERE is_active = true;

ROLLBACK; -- Don't commit the test transaction

-- =============================================================================
-- VALIDATION RESULTS INTERPRETATION
-- =============================================================================

-- Expected Results:
-- 1. Keyword relationships table should have 50+ relationship mappings
-- 2. Keyword contexts table should have 100+ context mappings
-- 3. Technology keywords should expand to include synonyms like 'tech', 'digital', 'innovation'
-- 4. Healthcare keywords should expand to include 'clinical', 'medical', 'patient care'
-- 5. Financial keywords should expand to include 'finance', 'credit', 'lending'
-- 6. Total unique keywords should exceed 2000
-- 7. Coverage percentage should be >80% (keywords with relationships)
-- 8. Average confidence scores should be >0.75
-- 9. Query performance should be <100ms for keyword expansion
-- 10. Sample business classification should show multiple matching industries

-- Success Criteria:
-- ✅ 2000+ unique keywords across all industries
-- ✅ 50+ keyword relationships with >75% average confidence
-- ✅ 100+ keyword contexts with appropriate weights
-- ✅ Cross-industry keyword sharing for common terms
-- ✅ High-performance keyword expansion queries
-- ✅ Accurate business classification with expanded keywords
