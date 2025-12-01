-- =====================================================
-- Expand Keyword Population - Phase 1
-- Purpose: Expand keywords to 90%+ of codes having 15+ keywords
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 1, Task 1.2
-- =====================================================
-- 
-- This script extracts keywords from code_metadata official descriptions
-- and populates the code_keywords table by matching codes to classification_codes.
-- Target: 90%+ of codes have 15+ keywords with average 18 keywords per code.
-- =====================================================

-- =====================================================
-- Part 1: Function to Extract Keywords from Text
-- =====================================================

-- Create or replace function to extract keywords from text
CREATE OR REPLACE FUNCTION extract_keywords_from_description(description_text TEXT)
RETURNS TEXT[] AS $$
DECLARE
    keywords TEXT[];
    words TEXT[];
    word TEXT;
    stop_words TEXT[] := ARRAY[
        'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
        'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
        'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
        'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
        'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
        'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
        'services', 'service', 'products', 'product', 'business', 'company',
        'companies', 'other', 'etc', 'including', 'excluding', 'related', 'not',
        'primarily', 'engaged', 'establishments', 'establishment', 'comprises',
        'industry', 'industries', 'u.s.', 'us', 'providing', 'provides'
    ];
BEGIN
    -- Convert to lowercase and remove punctuation
    description_text := lower(regexp_replace(description_text, '[^a-z0-9\s]', ' ', 'g'));
    
    -- Split into words
    words := string_to_array(description_text, ' ');
    
    -- Filter words
    keywords := ARRAY[]::TEXT[];
    FOREACH word IN ARRAY words
    LOOP
        word := trim(word);
        
        -- Skip empty, short words, stop words, and numbers
        IF word != '' 
           AND length(word) >= 3 
           AND word != ALL(stop_words)
           AND word !~ '^\d+$' THEN
            keywords := array_append(keywords, word);
        END IF;
    END LOOP;
    
    RETURN keywords;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- =====================================================
-- Part 2: Extract Keywords from code_metadata and Link to classification_codes
-- =====================================================

-- Insert keywords extracted from code_metadata official descriptions
-- Match codes by code_type and code to find classification_codes.id
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    keyword,
    CASE
        -- Higher relevance for official codes
        WHEN cm.is_official THEN 0.95
        -- Higher relevance for active codes
        WHEN cm.is_active THEN 0.90
        ELSE 0.85
    END as relevance_score,
    'exact' as match_type
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN LATERAL unnest(extract_keywords_from_description(cm.official_description)) as keyword
WHERE cm.is_active = true
  AND cm.official_description IS NOT NULL
  AND cm.official_description != ''
  AND keyword IS NOT NULL
  AND length(keyword) >= 3
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score),
    match_type = CASE WHEN EXCLUDED.match_type = 'exact' THEN 'exact' ELSE code_keywords.match_type END;

-- =====================================================
-- Part 3: Add Industry-Specific Keywords Based on Industry Mappings
-- =====================================================

-- Technology Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('software'), ('technology'), ('tech'), ('digital'), ('computer'), ('programming'),
    ('development'), ('coding'), ('app'), ('application'), ('platform'), ('system'),
    ('saas'), ('cloud'), ('api'), ('web'), ('mobile'), ('ios'), ('android'),
    ('devops'), ('infrastructure'), ('server'), ('database'), ('backend'), ('frontend'),
    ('code'), ('developer'), ('engineer'), ('it'), ('information'), ('data'),
    ('automation'), ('integration'), ('solution'), ('service'), ('tool'), ('framework')
) AS tech_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Technology'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%technology%'
       OR 'Technology' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || tech_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Healthcare Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('healthcare'), ('medical'), ('health'), ('hospital'), ('clinic'), ('doctor'),
    ('physician'), ('patient'), ('medicine'), ('pharmaceutical'), ('drug'), ('therapy'),
    ('treatment'), ('diagnosis'), ('surgery'), ('nursing'), ('wellness'), ('fitness'),
    ('dental'), ('vision'), ('mental'), ('psychology'), ('counseling'), ('rehabilitation'),
    ('emergency'), ('ambulance'), ('pharmacy'), ('laboratory'), ('radiology'), ('imaging')
) AS health_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Healthcare'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%healthcare%'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%medical%'
       OR 'Healthcare' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || health_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Financial Services Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('finance'), ('financial'), ('banking'), ('bank'), ('investment'), ('investing'),
    ('trading'), ('stock'), ('market'), ('fund'), ('capital'), ('credit'), ('loan'),
    ('mortgage'), ('insurance'), ('payment'), ('transaction'), ('card'), ('debit'),
    ('accounting'), ('accountant'), ('audit'), ('tax'), ('account'), ('savings'),
    ('checking'), ('wealth'), ('asset'), ('portfolio'), ('broker'), ('advisor')
) AS finance_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Financial Services'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%financial%'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%finance%'
       OR 'Financial Services' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || finance_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Retail & Commerce Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('retail'), ('store'), ('shop'), ('shopping'), ('merchant'), ('commerce'),
    ('ecommerce'), ('online'), ('marketplace'), ('product'), ('goods'), ('sale'),
    ('selling'), ('buying'), ('purchase'), ('customer'), ('consumer'), ('grocery'),
    ('supermarket'), ('department'), ('boutique'), ('outlet'), ('warehouse'), ('wholesale'),
    ('fashion'), ('apparel'), ('clothing'), ('accessory'), ('jewelry'), ('footwear')
) AS retail_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Retail & Commerce'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%retail%'
       OR 'Retail & Commerce' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || retail_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Food & Beverage Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('restaurant'), ('food'), ('dining'), ('cafe'), ('bistro'), ('bar'), ('grill'),
    ('catering'), ('bakery'), ('pizza'), ('burger'), ('cuisine'), ('menu'), ('chef'),
    ('kitchen'), ('beverage'), ('drink'), ('coffee'), ('tea'), ('alcohol'), ('wine'),
    ('beer'), ('liquor'), ('dining'), ('meal'), ('breakfast'), ('lunch'), ('dinner')
) AS food_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Food & Beverage'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%food%'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%beverage%'
       OR 'Food & Beverage' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || food_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Manufacturing Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('manufacturing'), ('production'), ('factory'), ('industrial'), ('machinery'),
    ('equipment'), ('assembly'), ('fabrication'), ('processing'), ('machining'),
    ('welding'), ('cutting'), ('molding'), ('casting'), ('packaging'), ('quality'),
    ('control'), ('automation'), ('robotics'), ('cnc'), ('metal'), ('plastic'),
    ('wood'), ('textile'), ('chemical'), ('material'), ('component'), ('part')
) AS manufacturing_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Manufacturing'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%manufacturing%'
       OR 'Manufacturing' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || manufacturing_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Construction Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('construction'), ('building'), ('contractor'), ('development'), ('building'),
    ('renovation'), ('remodeling'), ('architecture'), ('engineering'), ('contracting'),
    ('residential'), ('commercial'), ('institutional'), ('industrial'), ('infrastructure'),
    ('foundation'), ('framing'), ('masonry'), ('electrical'), ('plumbing'), ('hvac')
) AS construction_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Construction'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%construction%'
       OR 'Construction' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || construction_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Transportation Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('transportation'), ('transport'), ('logistics'), ('shipping'), ('delivery'),
    ('freight'), ('cargo'), ('trucking'), ('truck'), ('vehicle'), ('fleet'),
    ('warehouse'), ('warehousing'), ('distribution'), ('supply'), ('chain'), ('storage'),
    ('courier'), ('express'), ('parcel'), ('package'), ('mail'), ('postal'), ('airline'),
    ('railway'), ('railroad'), ('transit'), ('passenger'), ('freight')
) AS transport_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Transportation'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%transportation%'
       OR 'Transportation' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || transport_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Education Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('education'), ('school'), ('university'), ('college'), ('learning'), ('teaching'),
    ('training'), ('course'), ('class'), ('student'), ('teacher'), ('instructor'),
    ('academic'), ('curriculum'), ('degree'), ('certificate'), ('diploma'), ('tutoring'),
    ('online'), ('distance'), ('e-learning'), ('edtech'), ('educational'), ('institution')
) AS education_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Education'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%education%'
       OR 'Education' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || education_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Hospitality Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('hospitality'), ('hotel'), ('lodging'), ('accommodation'), ('resort'),
    ('tourism'), ('travel'), ('vacation'), ('hospitality'), ('motel'), ('inn'),
    ('bed'), ('breakfast'), ('hostel'), ('guest'), ('room'), ('suite'), ('conference')
) AS hospitality_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Hospitality'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%hospitality%'
       OR 'Hospitality' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || hospitality_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Professional Services Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('professional'), ('business'), ('corporate'), ('commercial'), ('enterprise'),
    ('consulting'), ('advisory'), ('services'), ('management'), ('legal'), ('law'),
    ('attorney'), ('lawyer'), ('accounting'), ('audit'), ('tax'), ('financial'), ('advisory')
) AS professional_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Professional Services'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%professional%'
       OR 'Professional Services' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || professional_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Real Estate Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('real estate'), ('property'), ('realty'), ('land'), ('housing'),
    ('rental'), ('leasing'), ('brokerage'), ('development'), ('residential'),
    ('commercial'), ('apartment'), ('condo'), ('house'), ('building'), ('estate')
) AS realestate_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Real Estate'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%real estate%'
       OR 'Real Estate' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || realestate_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Arts and Entertainment Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM code_metadata cm
INNER JOIN classification_codes cc 
    ON cc.code_type = cm.code_type 
    AND cc.code = cm.code
    AND cc.is_active = true
CROSS JOIN (VALUES
    ('arts'), ('entertainment'), ('theater'), ('music'), ('performance'),
    ('cultural'), ('creative'), ('artistic'), ('amusement'), ('park'), ('theme'),
    ('museum'), ('gallery'), ('exhibition'), ('show'), ('concert'), ('festival')
) AS arts_keywords(keyword)
WHERE cm.is_active = true
  AND (cm.industry_mappings->>'primary_industry' = 'Arts and Entertainment'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%arts%'
       OR cm.industry_mappings->>'primary_industry' ILIKE '%entertainment%'
       OR 'Arts and Entertainment' = ANY(SELECT jsonb_array_elements_text(cm.industry_mappings->'secondary_industries')))
  AND lower(cm.official_description) LIKE '%' || arts_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Part 4: Add Synonym Keywords
-- =====================================================

-- Add synonym keywords based on existing keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    ck1.code_id,
    synonym,
    0.85 as relevance_score, -- Slightly lower relevance for synonyms
    'synonym' as match_type
FROM code_keywords ck1
CROSS JOIN (VALUES
    ('software', 'app'), ('software', 'application'), ('software', 'program'),
    ('technology', 'tech'), ('technology', 'it'), ('technology', 'digital'),
    ('computer', 'pc'), ('computer', 'machine'), ('computer', 'system'),
    ('programming', 'coding'), ('programming', 'development'), ('programming', 'engineering'),
    ('healthcare', 'medical'), ('healthcare', 'health'), ('healthcare', 'clinical'),
    ('doctor', 'physician'), ('doctor', 'practitioner'), ('doctor', 'clinician'),
    ('hospital', 'medical'), ('hospital', 'clinic'), ('hospital', 'facility'),
    ('retail', 'store'), ('retail', 'shop'), ('retail', 'commerce'),
    ('restaurant', 'cafe'), ('restaurant', 'dining'), ('restaurant', 'eatery'),
    ('banking', 'bank'), ('banking', 'financial'), ('banking', 'finance'),
    ('education', 'school'), ('education', 'learning'), ('education', 'academic'),
    ('construction', 'building'), ('construction', 'contractor'), ('construction', 'development')
) AS synonyms(keyword, synonym)
WHERE ck1.keyword = synonyms.keyword
  AND ck1.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Part 5: Verification Queries
-- =====================================================

-- Count keywords per code
SELECT 
    'Keywords per Code Statistics' AS metric,
    COUNT(DISTINCT code_id) AS codes_with_keywords,
    COUNT(*) AS total_keywords,
    ROUND(AVG(keyword_count), 2) AS avg_keywords_per_code,
    MIN(keyword_count) AS min_keywords,
    MAX(keyword_count) AS max_keywords
FROM (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts;

-- Codes with 15+ keywords
SELECT 
    'Codes with 15+ Keywords' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM classification_codes WHERE is_active = true), 2) AS percentage
FROM (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
    HAVING COUNT(*) >= 15
) AS codes_with_15_plus;

-- Codes with less than 15 keywords (needs improvement)
SELECT 
    'Codes with < 15 Keywords' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM classification_codes WHERE is_active = true), 2) AS percentage
FROM (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
    HAVING COUNT(*) < 15
) AS codes_with_less_than_15;

-- Keyword coverage by code type
SELECT 
    'Keyword Coverage by Code Type' AS metric,
    cc.code_type,
    COUNT(DISTINCT ck.code_id) AS codes_with_keywords,
    COUNT(DISTINCT cc.id) AS total_codes,
    ROUND(COUNT(DISTINCT ck.code_id) * 100.0 / COUNT(DISTINCT cc.id), 2) AS coverage_percentage,
    ROUND(AVG(keyword_counts.keyword_count), 2) AS avg_keywords_per_code
FROM classification_codes cc
LEFT JOIN code_keywords ck ON ck.code_id = cc.id
LEFT JOIN (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts ON keyword_counts.code_id = cc.id
WHERE cc.is_active = true
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- Summary report
SELECT 
    '=== KEYWORD EXPANSION SUMMARY ===' AS section;

SELECT 
    'Total Codes' AS metric,
    COUNT(*) AS value
FROM classification_codes
WHERE is_active = true

UNION ALL

SELECT 
    'Codes with Keywords' AS metric,
    COUNT(DISTINCT code_id) AS value
FROM code_keywords

UNION ALL

SELECT 
    'Codes with 15+ Keywords' AS metric,
    COUNT(*) AS value
FROM (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
    HAVING COUNT(*) >= 15
) AS codes_with_15_plus

UNION ALL

SELECT 
    'Average Keywords per Code' AS metric,
    ROUND(AVG(keyword_count), 2) AS value
FROM (
    SELECT code_id, COUNT(*) as keyword_count
    FROM code_keywords
    GROUP BY code_id
) AS keyword_counts

UNION ALL

SELECT 
    'Keyword Coverage Percentage' AS metric,
    ROUND(COUNT(DISTINCT code_id) * 100.0 / (SELECT COUNT(*) FROM classification_codes WHERE is_active = true), 2) AS value
FROM code_keywords;

