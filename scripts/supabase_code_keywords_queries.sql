-- =====================================================
-- Useful Queries for code_keywords Table
-- Run these in Supabase SQL Editor to explore your data
-- =====================================================

-- =====================================================
-- 1. Overview Statistics
-- =====================================================

-- Total keyword mappings
SELECT 
    'Total Keywords' as metric,
    COUNT(*) as value
FROM code_keywords;

-- Keywords by match type
SELECT 
    match_type,
    COUNT(*) as count,
    ROUND(AVG(relevance_score), 2) as avg_relevance,
    ROUND(MIN(relevance_score), 2) as min_relevance,
    ROUND(MAX(relevance_score), 2) as max_relevance
FROM code_keywords
GROUP BY match_type
ORDER BY match_type;

-- Keywords by code type
SELECT 
    cc.code_type,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- =====================================================
-- 2. Search for Specific Keywords
-- =====================================================

-- Find all codes for a specific keyword (e.g., "software")
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) = 'software'
ORDER BY ck.relevance_score DESC, cc.code_type, cc.code
LIMIT 20;

-- Find codes matching multiple keywords
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    STRING_AGG(DISTINCT ck.keyword, ', ') as matched_keywords,
    COUNT(DISTINCT ck.keyword) as keyword_count,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) IN ('software', 'technology', 'platform')
  AND ck.relevance_score >= 0.50
GROUP BY cc.id, cc.code_type, cc.code, cc.description
ORDER BY keyword_count DESC, avg_relevance DESC
LIMIT 20;

-- =====================================================
-- 3. Search by Code Type
-- =====================================================

-- MCC codes with keywords
SELECT 
    cc.code,
    cc.description,
    COUNT(ck.id) as keyword_count,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'MCC'
GROUP BY cc.id, cc.code, cc.description
ORDER BY keyword_count DESC, avg_relevance DESC
LIMIT 20;

-- NAICS codes with keywords
SELECT 
    cc.code,
    cc.description,
    COUNT(ck.id) as keyword_count,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'NAICS'
GROUP BY cc.id, cc.code, cc.description
ORDER BY keyword_count DESC, avg_relevance DESC
LIMIT 20;

-- SIC codes with keywords
SELECT 
    cc.code,
    cc.description,
    COUNT(ck.id) as keyword_count,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE cc.code_type = 'SIC'
GROUP BY cc.id, cc.code, cc.description
ORDER BY keyword_count DESC, avg_relevance DESC
LIMIT 20;

-- =====================================================
-- 4. Search by Industry
-- =====================================================

-- Keywords for Technology industry codes
SELECT 
    i.name as industry,
    cc.code_type,
    cc.code,
    cc.description,
    COUNT(ck.id) as keyword_count,
    STRING_AGG(DISTINCT ck.keyword, ', ' ORDER BY ck.keyword) as keywords
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name = 'Technology'
GROUP BY i.name, cc.id, cc.code_type, cc.code, cc.description
ORDER BY keyword_count DESC
LIMIT 20;

-- =====================================================
-- 5. High Relevance Keywords
-- =====================================================

-- Top keywords by relevance score
SELECT 
    ck.keyword,
    COUNT(*) as code_count,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance,
    ROUND(MAX(ck.relevance_score), 2) as max_relevance,
    ck.match_type
FROM code_keywords ck
GROUP BY ck.keyword, ck.match_type
HAVING AVG(ck.relevance_score) >= 0.90
ORDER BY avg_relevance DESC, code_count DESC
LIMIT 30;

-- =====================================================
-- 6. Synonym Analysis
-- =====================================================

-- Find synonyms for a keyword
SELECT 
    ck1.keyword as original_keyword,
    ck2.keyword as synonym,
    ck2.relevance_score,
    cc.code_type,
    cc.code,
    cc.description
FROM code_keywords ck1
INNER JOIN code_keywords ck2 ON ck1.code_id = ck2.code_id
INNER JOIN classification_codes cc ON cc.id = ck1.code_id
WHERE ck1.match_type = 'exact'
  AND ck2.match_type = 'synonym'
  AND lower(ck1.keyword) = 'software'
ORDER BY ck2.relevance_score DESC
LIMIT 20;

-- =====================================================
-- 7. Code Coverage Analysis
-- =====================================================

-- Codes with most keywords
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    COUNT(ck.id) as keyword_count,
    STRING_AGG(DISTINCT ck.match_type, ', ') as match_types,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
GROUP BY cc.id, cc.code_type, cc.code, cc.description
ORDER BY keyword_count DESC
LIMIT 20;

-- Codes with no keywords (missing coverage)
SELECT 
    cc.code_type,
    cc.code,
    cc.description
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code_id = cc.id
WHERE ck.id IS NULL
  AND cc.is_active = true
ORDER BY cc.code_type, cc.code
LIMIT 50;

-- =====================================================
-- 8. Test Keyword Lookup (Simulate API Query)
-- =====================================================

-- Simulate what the API does: find codes for keywords
-- This matches the logic in GetClassificationCodesByKeywords
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score,
    ck.match_type
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) IN ('software', 'technology', 'platform')
  AND cc.code_type = 'MCC'
  AND ck.relevance_score >= 0.50
ORDER BY ck.relevance_score DESC, cc.code
LIMIT 20;

-- =====================================================
-- 9. Industry Keyword Coverage
-- =====================================================

-- Keywords per industry
SELECT 
    i.name as industry,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    COUNT(DISTINCT ck.keyword) as unique_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
GROUP BY i.name
ORDER BY total_keywords DESC;

-- =====================================================
-- 10. Quality Checks
-- =====================================================

-- Check for duplicate keyword entries (should be none due to UNIQUE constraint)
SELECT 
    code_id,
    keyword,
    COUNT(*) as duplicate_count
FROM code_keywords
GROUP BY code_id, keyword
HAVING COUNT(*) > 1;

-- Check for keywords with invalid relevance scores
SELECT 
    id,
    code_id,
    keyword,
    relevance_score
FROM code_keywords
WHERE relevance_score < 0.00 OR relevance_score > 1.00;

-- Check for keywords with invalid match types
SELECT 
    id,
    code_id,
    keyword,
    match_type
FROM code_keywords
WHERE match_type NOT IN ('exact', 'partial', 'synonym');

