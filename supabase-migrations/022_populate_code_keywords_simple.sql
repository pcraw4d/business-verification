-- Simplified Migration: Populate code_keywords table
-- This is a simpler, more reliable version that directly inserts keywords

-- =====================================================
-- Step 1: Extract keywords from classification code descriptions
-- =====================================================

-- Insert keywords extracted from code descriptions
-- This uses a simple approach: split description into words and filter
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    trim(lower(keyword)) as keyword,
    CASE
        WHEN cc.is_primary THEN 0.95
        WHEN cc.confidence >= 0.90 THEN 0.90
        WHEN cc.confidence >= 0.80 THEN 0.85
        WHEN cc.confidence >= 0.70 THEN 0.75
        ELSE 0.70
    END as relevance_score,
    'exact' as match_type
FROM classification_codes cc,
LATERAL unnest(
    string_to_array(
        regexp_replace(
            lower(cc.description),
            '[^a-z0-9\s]',
            ' ',
            'g'
        ),
        ' '
    )
) as keyword
WHERE trim(keyword) != ''
  AND length(trim(keyword)) >= 3
  AND trim(keyword) NOT IN (
    'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
    'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
    'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
    'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
    'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
    'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
    'services', 'service', 'products', 'product', 'business', 'company',
    'companies', 'other', 'etc', 'including', 'excluding', 'related', 'not'
  )
  AND trim(keyword) !~ '^\d+$' -- Exclude pure numbers
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 2: Link industry keywords to classification codes
-- =====================================================

-- Link industry keywords to codes in the same industry
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    ik.keyword,
    LEAST(0.95, (ik.weight * 0.7 + cc.confidence * 0.3)) as relevance_score,
    CASE WHEN ik.is_primary THEN 'exact' ELSE 'partial' END as match_type
FROM classification_codes cc
INNER JOIN industry_keywords ik ON ik.industry_id = cc.industry_id
WHERE lower(cc.description) LIKE '%' || lower(ik.keyword) || '%'
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score),
    match_type = CASE WHEN EXCLUDED.match_type = 'exact' THEN 'exact' ELSE code_keywords.match_type END;

-- =====================================================
-- Step 3: Add common synonyms
-- =====================================================

-- Add common business term synonyms
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    ck.code_id,
    synonym,
    ck.relevance_score * 0.85,
    'synonym' as match_type
FROM code_keywords ck
CROSS JOIN (VALUES
    ('software', 'app'), ('software', 'application'), ('software', 'program'),
    ('technology', 'tech'), ('technology', 'digital'),
    ('retail', 'store'), ('retail', 'shop'),
    ('restaurant', 'food'), ('restaurant', 'dining'),
    ('healthcare', 'medical'), ('healthcare', 'health'),
    ('finance', 'financial'), ('finance', 'banking')
) AS synonyms(original, synonym)
WHERE ck.keyword = synonyms.original
  AND ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 4: Display statistics
-- =====================================================

SELECT 
    'Code Keywords Population Complete' as status,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT code_id) as codes_with_keywords,
    ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) as avg_keywords_per_code
FROM code_keywords;

