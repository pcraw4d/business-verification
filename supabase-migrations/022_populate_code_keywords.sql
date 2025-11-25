-- Migration: Populate code_keywords table for hybrid code generation
-- This migration extracts keywords from classification code descriptions and
-- links industry keywords to their related classification codes

-- =====================================================
-- 1. Extract keywords from classification code descriptions
-- =====================================================

-- Function to extract keywords from text (removes common stop words)
CREATE OR REPLACE FUNCTION extract_keywords_from_text(text_content TEXT)
RETURNS TEXT[] AS $$
DECLARE
    words TEXT[];
    filtered_words TEXT[];
    word TEXT;
    stop_words TEXT[] := ARRAY[
        'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
        'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
        'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
        'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
        'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
        'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
        'services', 'service', 'products', 'product', 'business', 'company',
        'companies', 'other', 'etc', 'including', 'excluding', 'related'
    ];
BEGIN
    -- Convert to lowercase and split into words
    text_content := lower(trim(text_content));
    
    -- Remove special characters and split
    text_content := regexp_replace(text_content, '[^a-z0-9\s]', ' ', 'g');
    words := string_to_array(text_content, ' ');
    
    -- Filter out stop words, empty strings, and short words (< 3 chars)
    filtered_words := ARRAY[]::TEXT[];
    FOREACH word IN ARRAY words
    LOOP
        word := trim(word);
        IF length(word) >= 3 
           AND word != ALL(stop_words)
           AND word !~ '^\d+$' -- Exclude pure numbers
        THEN
            filtered_words := array_append(filtered_words, word);
        END IF;
    END LOOP;
    
    RETURN filtered_words;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- =====================================================
-- 2. Populate code_keywords from classification code descriptions
-- =====================================================

-- Insert keywords extracted from code descriptions
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    CASE
        -- Higher relevance for primary codes
        WHEN cc.is_primary THEN 0.95
        -- Higher relevance for codes with high confidence
        WHEN cc.confidence >= 0.90 THEN 0.90
        WHEN cc.confidence >= 0.80 THEN 0.85
        WHEN cc.confidence >= 0.70 THEN 0.75
        ELSE 0.70
    END as relevance_score,
    'exact' as match_type
FROM classification_codes cc
CROSS JOIN LATERAL unnest(extract_keywords_from_text(cc.description)) as keyword
WHERE keyword IS NOT NULL
  AND length(keyword) >= 3
  AND keyword NOT IN ('and', 'the', 'for', 'with', 'from', 'that', 'this')
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- 3. Link industry keywords to their classification codes
-- =====================================================

-- Insert keywords from industry_keywords table that match code descriptions
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    ik.keyword,
    -- Relevance based on industry keyword weight and code confidence
    LEAST(0.95, (ik.weight * 0.7 + cc.confidence * 0.3)) as relevance_score,
    CASE
        WHEN ik.is_primary THEN 'exact'
        ELSE 'partial'
    END as match_type
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
INNER JOIN industry_keywords ik ON ik.industry_id = i.id
WHERE lower(cc.description) LIKE '%' || lower(ik.keyword) || '%'
  OR lower(ik.keyword) = ANY(extract_keywords_from_text(cc.description))
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(
        code_keywords.relevance_score,
        EXCLUDED.relevance_score
    ),
    match_type = CASE
        WHEN EXCLUDED.match_type = 'exact' THEN 'exact'
        ELSE code_keywords.match_type
    END;

-- =====================================================
-- 4. Add common business term synonyms
-- =====================================================

-- Add common synonyms and variations for better matching
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    ck.code_id,
    synonym.keyword,
    ck.relevance_score * 0.85 as relevance_score, -- Slightly lower for synonyms
    'synonym' as match_type
FROM code_keywords ck
CROSS JOIN LATERAL (
    SELECT keyword FROM (VALUES
        ('software', 'app'), ('software', 'application'), ('software', 'program'),
        ('software', 'platform'), ('software', 'system'),
        ('technology', 'tech'), ('technology', 'digital'), ('technology', 'it'),
        ('retail', 'store'), ('retail', 'shop'), ('retail', 'merchant'),
        ('restaurant', 'food'), ('restaurant', 'dining'), ('restaurant', 'cafe'),
        ('healthcare', 'medical'), ('healthcare', 'health'), ('healthcare', 'hospital'),
        ('finance', 'financial'), ('finance', 'banking'), ('finance', 'investment'),
        ('education', 'school'), ('education', 'learning'), ('education', 'training'),
        ('manufacturing', 'production'), ('manufacturing', 'factory'), ('manufacturing', 'industrial')
    ) AS synonyms(original, keyword)
    WHERE synonyms.original = ck.keyword
) synonym
WHERE ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- 5. Add code-specific keywords based on code patterns
-- =====================================================

-- For MCC codes: add common merchant category keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    0.80 as relevance_score,
    'partial' as match_type
FROM classification_codes cc
CROSS JOIN LATERAL (
    SELECT keyword FROM (VALUES
        ('merchant'), ('payment'), ('transaction'), ('card'), ('credit'), ('debit'),
        ('point'), ('sale'), ('pos'), ('terminal'), ('processing')
    ) AS mcc_keywords(keyword)
) keywords
WHERE cc.code_type = 'MCC'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- For NAICS codes: add industry sector keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    0.85 as relevance_score,
    'partial' as match_type
FROM classification_codes cc
CROSS JOIN LATERAL (
    SELECT keyword FROM (VALUES
        ('industry'), ('sector'), ('establishment'), ('enterprise'), ('business'),
        ('activity'), ('operation'), ('production'), ('manufacturing'), ('service')
    ) AS naics_keywords(keyword)
) keywords
WHERE cc.code_type = 'NAICS'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- For SIC codes: add standard industry keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    0.80 as relevance_score,
    'partial' as match_type
FROM classification_codes cc
CROSS JOIN LATERAL (
    SELECT keyword FROM (VALUES
        ('standard'), ('industry'), ('classification'), ('division'), ('major'),
        ('group'), ('establishment'), ('activity')
    ) AS sic_keywords(keyword)
) keywords
WHERE cc.code_type = 'SIC'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- 6. Statistics and verification
-- =====================================================

-- Display statistics
DO $$
DECLARE
    total_keywords INTEGER;
    total_codes INTEGER;
    avg_keywords_per_code NUMERIC;
BEGIN
    SELECT COUNT(*) INTO total_keywords FROM code_keywords;
    SELECT COUNT(DISTINCT code_id) INTO total_codes FROM code_keywords;
    SELECT ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) 
           INTO avg_keywords_per_code 
           FROM code_keywords;
    
    RAISE NOTICE 'Code Keywords Population Complete:';
    RAISE NOTICE '  Total keywords: %', total_keywords;
    RAISE NOTICE '  Codes with keywords: %', total_codes;
    RAISE NOTICE '  Average keywords per code: %', avg_keywords_per_code;
END $$;

-- Create index for faster keyword lookups (if not already exists)
CREATE INDEX IF NOT EXISTS idx_code_keywords_keyword_lower ON code_keywords(lower(keyword));
CREATE INDEX IF NOT EXISTS idx_code_keywords_relevance_code ON code_keywords(relevance_score DESC, code_id);

-- Add comment
COMMENT ON FUNCTION extract_keywords_from_text IS 'Extracts meaningful keywords from text by removing stop words and normalizing';

