-- =====================================================
-- Enhanced Keyword Coverage Verification
-- Purpose: Verify keyword coverage meets Phase 1, Task 1.2 goals
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 1, Task 1.2
-- =====================================================
-- 
-- Targets:
-- - 90%+ of codes have 15+ keywords
-- - Average 18 keywords per code
-- - Keyword matching accuracy > 85%
-- =====================================================

-- =====================================================
-- Part 1: Basic Statistics
-- =====================================================

-- Total keywords
SELECT 
    'Total Keywords' AS metric,
    COUNT(*) AS value
FROM code_keywords;

-- Total codes with keywords
SELECT 
    'Total Codes with Keywords' AS metric,
    COUNT(DISTINCT code_id) AS value,
    ROUND(COUNT(DISTINCT code_id) * 100.0 / (SELECT COUNT(*) FROM classification_codes), 2) AS percentage
FROM code_keywords;

-- =====================================================
-- Part 2: Keywords per Code Type
-- =====================================================

SELECT 
    'Keywords per Code Type' AS metric,
    cc.code_type,
    COUNT(DISTINCT ck.code_id) AS codes_with_keywords,
    COUNT(ck.id) AS total_keywords,
    ROUND(AVG(ck.relevance_score), 2) AS avg_relevance,
    ROUND(COUNT(ck.id)::numeric / NULLIF(COUNT(DISTINCT ck.code_id), 0), 2) AS avg_keywords_per_code
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- =====================================================
-- Part 3: Keyword Coverage by Code (15+ Target)
-- =====================================================

-- Codes with 15+ keywords
SELECT 
    'Codes with 15+ Keywords' AS metric,
    cc.code_type,
    COUNT(DISTINCT ck.code_id) AS codes_with_15plus_keywords,
    COUNT(DISTINCT cc.id) AS total_codes,
    ROUND(COUNT(DISTINCT ck.code_id) * 100.0 / NULLIF(COUNT(DISTINCT cc.id), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(DISTINCT ck.code_id) * 100.0 / NULLIF(COUNT(DISTINCT cc.id), 0) >= 90.0 THEN '✅ PASS - 90%+ coverage'
        ELSE '❌ FAIL - Below 90% coverage'
    END AS status
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE (
    SELECT COUNT(*) 
    FROM code_keywords ck2 
    WHERE ck2.code_id = ck.code_id
) >= 15
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- Overall coverage (all code types combined)
SELECT 
    'Overall Codes with 15+ Keywords' AS metric,
    COUNT(DISTINCT ck.code_id) AS codes_with_15plus_keywords,
    (SELECT COUNT(*) FROM classification_codes) AS total_codes,
    ROUND(COUNT(DISTINCT ck.code_id) * 100.0 / NULLIF((SELECT COUNT(*) FROM classification_codes), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(DISTINCT ck.code_id) * 100.0 / NULLIF((SELECT COUNT(*) FROM classification_codes), 0) >= 90.0 THEN '✅ PASS - 90%+ coverage'
        ELSE '❌ FAIL - Below 90% coverage'
    END AS status
FROM code_keywords ck
WHERE (
    SELECT COUNT(*) 
    FROM code_keywords ck2 
    WHERE ck2.code_id = ck.code_id
) >= 15;

-- Distribution of keywords per code
SELECT 
    'Keyword Distribution' AS metric,
    keyword_range,
    COUNT(*) AS code_count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM (
    SELECT 
        code_id,
        COUNT(*) AS keyword_count,
        CASE 
            WHEN COUNT(*) >= 20 THEN '20+ keywords'
            WHEN COUNT(*) >= 15 THEN '15-19 keywords'
            WHEN COUNT(*) >= 10 THEN '10-14 keywords'
            WHEN COUNT(*) >= 5 THEN '5-9 keywords'
            ELSE '1-4 keywords'
        END AS keyword_range,
        CASE 
            WHEN COUNT(*) >= 20 THEN 1
            WHEN COUNT(*) >= 15 THEN 2
            WHEN COUNT(*) >= 10 THEN 3
            WHEN COUNT(*) >= 5 THEN 4
            ELSE 5
        END AS sort_order
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts
GROUP BY keyword_range, sort_order
ORDER BY sort_order;

-- Average keywords per code (target: 18)
SELECT 
    'Average Keywords per Code' AS metric,
    ROUND(AVG(keyword_count), 2) AS avg_keywords_per_code,
    CASE 
        WHEN AVG(keyword_count) >= 18.0 THEN '✅ PASS - Average 18+ keywords'
        ELSE '❌ FAIL - Below 18 keywords average'
    END AS status
FROM (
    SELECT 
        code_id,
        COUNT(*) AS keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts;

-- =====================================================
-- Part 4: Keywords by Match Type
-- =====================================================

SELECT 
    'Keywords by Match Type' AS metric,
    match_type,
    COUNT(*) AS count,
    ROUND(AVG(relevance_score), 2) AS avg_relevance,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM code_keywords
GROUP BY match_type
ORDER BY match_type;

-- =====================================================
-- Part 5: Sample Keywords for Quality Check
-- =====================================================

-- Sample keywords for MCC codes
SELECT 
    'Sample MCC Keywords' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'MCC'
ORDER BY ck.relevance_score DESC
LIMIT 20;

-- Sample keywords for NAICS codes
SELECT 
    'Sample NAICS Keywords' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'NAICS'
ORDER BY ck.relevance_score DESC
LIMIT 20;

-- Sample keywords for SIC codes
SELECT 
    'Sample SIC Keywords' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'SIC'
ORDER BY ck.relevance_score DESC
LIMIT 20;

-- =====================================================
-- Part 6: Keyword Lookup Test (Simulate API)
-- =====================================================

-- Test keyword lookup for common terms
SELECT 
    'Keyword Lookup Test - Software' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) IN ('software', 'technology', 'platform', 'development', 'programming')
  AND ck.relevance_score >= 0.50
ORDER BY ck.relevance_score DESC
LIMIT 10;

-- Test keyword lookup for healthcare terms
SELECT 
    'Keyword Lookup Test - Healthcare' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) IN ('medical', 'health', 'hospital', 'doctor', 'physician')
  AND ck.relevance_score >= 0.50
ORDER BY ck.relevance_score DESC
LIMIT 10;

-- =====================================================
-- Part 7: Codes Missing Keywords or Below Threshold
-- =====================================================

-- Codes with fewer than 15 keywords
SELECT 
    'Codes with < 15 Keywords' AS metric,
    cc.code_type,
    cc.code,
    cc.description,
    COUNT(ck.id) AS keyword_count
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code_id = cc.id
GROUP BY cc.code_type, cc.code, cc.description
HAVING COUNT(ck.id) < 15
ORDER BY cc.code_type, keyword_count ASC
LIMIT 50;

-- Codes with no keywords
SELECT 
    'Codes with No Keywords' AS metric,
    cc.code_type,
    cc.code,
    cc.description
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code_id = cc.id
WHERE ck.id IS NULL
ORDER BY cc.code_type, cc.code
LIMIT 50;

-- =====================================================
-- Part 8: Summary Report
-- =====================================================

SELECT 
    '=== KEYWORD COVERAGE SUMMARY ===' AS section;

SELECT 
    'Total Keywords' AS metric,
    COUNT(*) AS value,
    NULL AS status
FROM code_keywords

UNION ALL

SELECT 
    'Total Codes with Keywords' AS metric,
    COUNT(DISTINCT code_id) AS value,
    NULL AS status
FROM code_keywords

UNION ALL

SELECT 
    'Codes with 15+ Keywords' AS metric,
    COUNT(DISTINCT code_id) AS value,
    CASE 
        WHEN COUNT(DISTINCT code_id) * 100.0 / NULLIF((SELECT COUNT(*) FROM classification_codes), 0) >= 90.0 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_keywords
WHERE (
    SELECT COUNT(*) 
    FROM code_keywords ck2 
    WHERE ck2.code_id = code_keywords.code_id
) >= 15

UNION ALL

SELECT 
    'Average Keywords per Code' AS metric,
    ROUND(AVG(keyword_count), 2) AS value,
    CASE 
        WHEN AVG(keyword_count) >= 18.0 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM (
    SELECT 
        code_id,
        COUNT(*) AS keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts;

