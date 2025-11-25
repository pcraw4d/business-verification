-- Verification queries for code_keywords table

-- 1. Check total keywords
SELECT 
    'Total Keywords' as metric,
    COUNT(*) as value
FROM code_keywords;

-- 2. Check keywords per code type
SELECT 
    cc.code_type,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- 3. Check keywords by match type
SELECT 
    match_type,
    COUNT(*) as count,
    ROUND(AVG(relevance_score), 2) as avg_relevance
FROM code_keywords
GROUP BY match_type
ORDER BY match_type;

-- 4. Sample keywords for a specific code type
SELECT 
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

-- 5. Test keyword lookup (simulate what the API does)
SELECT 
    cc.code_type,
    cc.code,
    cc.description,
    ck.keyword,
    ck.relevance_score
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
WHERE lower(ck.keyword) IN ('software', 'technology', 'platform')
  AND cc.code_type = 'MCC'
  AND ck.relevance_score >= 0.50
ORDER BY ck.relevance_score DESC
LIMIT 10;

