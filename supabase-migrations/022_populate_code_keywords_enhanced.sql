-- Enhanced Migration: Populate code_keywords table with comprehensive keyword mappings
-- This migration includes:
-- 1. Keywords extracted from classification code descriptions
-- 2. Industry-specific keywords for all industries
-- 3. Common business terms and synonyms
-- 4. MCC, SIC, NAICS specific keywords
-- 5. Cross-industry keywords and variations

-- =====================================================
-- Step 1: Extract keywords from classification code descriptions
-- =====================================================

-- Insert keywords extracted from code descriptions
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    trim(lower(keyword)) as keyword,
    -- Use fixed relevance score since confidence column doesn't exist in actual table
    0.85 as relevance_score,
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
  AND trim(keyword) !~ '^\d+$'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 2: Link industry keywords to classification codes
-- =====================================================

-- Link industry keywords to codes in the same industry
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    ik.keyword,
    LEAST(0.95, (ik.weight * 0.8 + 0.15)) as relevance_score,
    CASE 
        WHEN ik.weight >= 0.90 THEN 'exact' 
        ELSE 'partial' 
    END as match_type
FROM classification_codes cc
INNER JOIN industry_keywords ik ON ik.industry_id = cc.industry_id
WHERE lower(cc.description) LIKE '%' || lower(ik.keyword) || '%'
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score),
    match_type = CASE WHEN EXCLUDED.match_type = 'exact' THEN 'exact' ELSE code_keywords.match_type END;

-- =====================================================
-- Step 3: Add comprehensive industry-specific keywords
-- =====================================================

-- Technology & Software Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('software'), ('technology'), ('tech'), ('digital'), ('computer'), ('programming'),
    ('development'), ('coding'), ('app'), ('application'), ('platform'), ('system'),
    ('saas'), ('cloud'), ('api'), ('web'), ('mobile'), ('ios'), ('android'),
    ('devops'), ('infrastructure'), ('server'), ('database'), ('backend'), ('frontend'),
    ('code'), ('developer'), ('engineer'), ('it'), ('information'), ('data'),
    ('automation'), ('integration'), ('solution'), ('service'), ('tool'), ('framework')
) AS tech_keywords(keyword)
WHERE i.name IN ('Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence')
  AND lower(cc.description) LIKE '%' || tech_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Healthcare & Medical Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('healthcare'), ('medical'), ('health'), ('hospital'), ('clinic'), ('doctor'),
    ('physician'), ('patient'), ('medicine'), ('pharmaceutical'), ('drug'), ('therapy'),
    ('treatment'), ('diagnosis'), ('surgery'), ('nursing'), ('wellness'), ('fitness'),
    ('dental'), ('vision'), ('mental'), ('psychology'), ('counseling'), ('rehabilitation'),
    ('emergency'), ('ambulance'), ('pharmacy'), ('laboratory'), ('radiology'), ('imaging')
) AS health_keywords(keyword)
WHERE i.name IN ('Healthcare', 'Medical Technology', 'Pharmaceuticals')
  AND lower(cc.description) LIKE '%' || health_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Finance & Banking Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('finance'), ('financial'), ('banking'), ('bank'), ('investment'), ('investing'),
    ('trading'), ('stock'), ('market'), ('fund'), ('capital'), ('credit'), ('loan'),
    ('mortgage'), ('insurance'), ('payment'), ('transaction'), ('card'), ('debit'),
    ('accounting'), ('accountant'), ('audit'), ('tax'), ('account'), ('savings'),
    ('checking'), ('wealth'), ('asset'), ('portfolio'), ('broker'), ('advisor')
) AS finance_keywords(keyword)
WHERE i.name IN ('Finance', 'Fintech', 'Insurance')
  AND lower(cc.description) LIKE '%' || finance_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Retail & E-commerce Industry Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('retail'), ('store'), ('shop'), ('shopping'), ('merchant'), ('commerce'),
    ('ecommerce'), ('online'), ('marketplace'), ('product'), ('goods'), ('sale'),
    ('selling'), ('buying'), ('purchase'), ('customer'), ('consumer'), ('grocery'),
    ('supermarket'), ('department'), ('boutique'), ('outlet'), ('warehouse'), ('wholesale'),
    ('fashion'), ('apparel'), ('clothing'), ('accessory'), ('jewelry'), ('footwear')
) AS retail_keywords(keyword)
WHERE i.name IN ('Retail', 'E-commerce', 'Fashion & Apparel')
  AND lower(cc.description) LIKE '%' || retail_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Manufacturing & Industrial Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('manufacturing'), ('production'), ('factory'), ('industrial'), ('machinery'),
    ('equipment'), ('assembly'), ('fabrication'), ('processing'), ('machining'),
    ('welding'), ('cutting'), ('molding'), ('casting'), ('packaging'), ('quality'),
    ('control'), ('automation'), ('robotics'), ('cnc'), ('metal'), ('plastic'),
    ('wood'), ('textile'), ('chemical'), ('material'), ('component'), ('part')
) AS manufacturing_keywords(keyword)
WHERE i.name IN ('Manufacturing', 'Industrial Technology', 'Construction')
  AND lower(cc.description) LIKE '%' || manufacturing_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Education & Training Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('education'), ('school'), ('university'), ('college'), ('learning'), ('teaching'),
    ('training'), ('course'), ('class'), ('student'), ('teacher'), ('instructor'),
    ('academic'), ('curriculum'), ('degree'), ('certificate'), ('diploma'), ('tutoring'),
    ('online'), ('distance'), ('e-learning'), ('edtech'), ('educational'), ('institution')
) AS education_keywords(keyword)
WHERE i.name IN ('Education', 'EdTech')
  AND lower(cc.description) LIKE '%' || education_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Transportation & Logistics Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('transportation'), ('transport'), ('logistics'), ('shipping'), ('delivery'),
    ('freight'), ('cargo'), ('trucking'), ('truck'), ('vehicle'), ('fleet'),
    ('warehouse'), ('warehousing'), ('distribution'), ('supply'), ('chain'), ('storage'),
    ('courier'), ('express'), ('parcel'), ('package'), ('mail'), ('postal')
) AS transport_keywords(keyword)
WHERE i.name IN ('Transportation', 'Logistics')
  AND lower(cc.description) LIKE '%' || transport_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Food & Beverage Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('restaurant'), ('food'), ('dining'), ('cafe'), ('bistro'), ('bar'), ('grill'),
    ('catering'), ('bakery'), ('pizza'), ('burger'), ('cuisine'), ('menu'), ('chef'),
    ('kitchen'), ('beverage'), ('drink'), ('coffee'), ('tea'), ('alcohol'), ('wine'),
    ('beer'), ('liquor'), ('dining'), ('meal'), ('breakfast'), ('lunch'), ('dinner')
) AS food_keywords(keyword)
WHERE i.name IN ('Food & Beverage', 'Food Technology')
  AND lower(cc.description) LIKE '%' || food_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Energy & Utilities Keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('energy'), ('power'), ('electric'), ('electricity'), ('utility'), ('utilities'),
    ('solar'), ('wind'), ('renewable'), ('sustainable'), ('green'), ('clean'),
    ('oil'), ('gas'), ('petroleum'), ('fuel'), ('nuclear'), ('hydro'), ('thermal'),
    ('generation'), ('distribution'), ('transmission'), ('grid'), ('battery'), ('storage')
) AS energy_keywords(keyword)
WHERE i.name IN ('Energy', 'Clean Energy')
  AND lower(cc.description) LIKE '%' || energy_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 4: Add comprehensive synonyms and variations
-- =====================================================

-- Technology synonyms
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
CROSS JOIN (VALUES
    ('software', 'app'), ('software', 'application'), ('software', 'program'), ('software', 'code'),
    ('technology', 'tech'), ('technology', 'digital'), ('technology', 'it'), ('technology', 'information'),
    ('development', 'dev'), ('development', 'programming'), ('development', 'coding'),
    ('platform', 'system'), ('platform', 'framework'), ('platform', 'solution'),
    ('cloud', 'hosting'), ('cloud', 'infrastructure'), ('cloud', 'saas'),
    ('data', 'information'), ('data', 'database'), ('data', 'analytics')
) AS tech_synonyms(original, synonym)
WHERE ck.keyword = tech_synonyms.original AND ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Healthcare synonyms
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
CROSS JOIN (VALUES
    ('healthcare', 'medical'), ('healthcare', 'health'), ('healthcare', 'hospital'),
    ('doctor', 'physician'), ('doctor', 'md'), ('doctor', 'practitioner'),
    ('medicine', 'drug'), ('medicine', 'pharmaceutical'), ('medicine', 'medication'),
    ('therapy', 'treatment'), ('therapy', 'rehabilitation'), ('therapy', 'counseling')
) AS health_synonyms(original, synonym)
WHERE ck.keyword = health_synonyms.original AND ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Finance synonyms
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
CROSS JOIN (VALUES
    ('finance', 'financial'), ('finance', 'banking'), ('finance', 'investment'),
    ('bank', 'banking'), ('bank', 'financial'), ('bank', 'institution'),
    ('payment', 'transaction'), ('payment', 'processing'), ('payment', 'card'),
    ('loan', 'credit'), ('loan', 'lending'), ('loan', 'mortgage'),
    ('insurance', 'coverage'), ('insurance', 'policy'), ('insurance', 'protection')
) AS finance_synonyms(original, synonym)
WHERE ck.keyword = finance_synonyms.original AND ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Retail synonyms
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT ck.code_id, synonym, ck.relevance_score * 0.85, 'synonym'
FROM code_keywords ck
CROSS JOIN (VALUES
    ('retail', 'store'), ('retail', 'shop'), ('retail', 'merchant'),
    ('ecommerce', 'online'), ('ecommerce', 'digital'), ('ecommerce', 'internet'),
    ('product', 'goods'), ('product', 'item'), ('product', 'merchandise'),
    ('sale', 'selling'), ('sale', 'transaction'), ('sale', 'purchase')
) AS retail_synonyms(original, synonym)
WHERE ck.keyword = retail_synonyms.original AND ck.match_type = 'exact'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 5: Add code-type-specific keywords
-- =====================================================

-- MCC (Merchant Category Code) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.80, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('merchant'), ('payment'), ('transaction'), ('card'), ('credit'), ('debit'),
    ('point'), ('sale'), ('pos'), ('terminal'), ('processing'), ('gateway'),
    ('processor'), ('acquirer'), ('issuer'), ('network'), ('visa'), ('mastercard'),
    ('amex'), ('discover'), ('chargeback'), ('refund'), ('authorization')
) AS mcc_keywords(keyword)
WHERE cc.code_type = 'MCC'
  AND lower(cc.description) LIKE '%' || mcc_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- NAICS (North American Industry Classification System) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.85, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('industry'), ('sector'), ('establishment'), ('enterprise'), ('business'),
    ('activity'), ('operation'), ('production'), ('manufacturing'), ('service'),
    ('classification'), ('category'), ('group'), ('division'), ('subsector')
) AS naics_keywords(keyword)
WHERE cc.code_type = 'NAICS'
  AND lower(cc.description) LIKE '%' || naics_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- SIC (Standard Industrial Classification) specific keywords
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.80, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('standard'), ('industry'), ('classification'), ('division'), ('major'),
    ('group'), ('establishment'), ('activity'), ('code'), ('category')
) AS sic_keywords(keyword)
WHERE cc.code_type = 'SIC'
  AND lower(cc.description) LIKE '%' || sic_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 6: Add common business terms
-- =====================================================

-- General business keywords that apply to many codes
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.75, 'partial'
FROM classification_codes cc
CROSS JOIN (VALUES
    ('business'), ('company'), ('corporation'), ('firm'), ('organization'),
    ('enterprise'), ('establishment'), ('operation'), ('service'), ('provider'),
    ('vendor'), ('supplier'), ('contractor'), ('consultant'), ('agency'),
    ('management'), ('administration'), ('support'), ('maintenance'), ('repair')
) AS business_keywords(keyword)
WHERE lower(cc.description) LIKE '%' || business_keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 7: Display statistics
-- =====================================================

SELECT 
    'Code Keywords Population Complete' as status,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT code_id) as codes_with_keywords,
    ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) as avg_keywords_per_code,
    COUNT(DISTINCT CASE WHEN match_type = 'exact' THEN code_id END) as codes_with_exact_matches,
    COUNT(DISTINCT CASE WHEN match_type = 'synonym' THEN code_id END) as codes_with_synonyms
FROM code_keywords;

-- Display breakdown by code type
SELECT 
    cc.code_type,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- Display breakdown by match type
SELECT 
    match_type,
    COUNT(*) as count,
    ROUND(AVG(relevance_score), 2) as avg_relevance,
    ROUND(MIN(relevance_score), 2) as min_relevance,
    ROUND(MAX(relevance_score), 2) as max_relevance
FROM code_keywords
GROUP BY match_type
ORDER BY match_type;

